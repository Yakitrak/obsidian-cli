package actions

import (
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// PropertySummary describes a vault property, note count, and inferred value details.
type PropertySummary struct {
	Name               string         `json:"name"`
	NoteCount          int            `json:"noteCount"`
	Shape              string         `json:"shape"`                // scalar, list, object, mixed, unknown
	ValueType          string         `json:"valueType"`            // bool, int, float, date, datetime, url, wikilink, string, object, mixed, unknown
	EnumValues         []string       `json:"enumValues,omitempty"` // enumerated values when small and suitable
	EnumValueCounts    map[string]int `json:"enumValueCounts,omitempty"`
	DistinctValueCount int            `json:"distinctValueCount"` // count of unique values encountered (lists flattened)
	TruncatedValueSet  bool           `json:"truncatedValueSet"`  // true if values exceeded maxValues cap
}

// PropertySource specifies which property sources to scan.
type PropertySource string

const (
	PropertySourceAll        PropertySource = "all"        // both frontmatter and inline
	PropertySourceFrontmatter PropertySource = "frontmatter" // YAML frontmatter only
	PropertySourceInline     PropertySource = "inline"     // dataview-style inline only (Key:: Value)
)

// PropertiesOptions controls scanning and enum detection behavior.
type PropertiesOptions struct {
	ExcludeTags        bool
	EnumThreshold      int // max distinct values to emit as enum
	MaxValues          int // cap stored values to avoid unbounded memory
	Notes              []string
	ForceEnumMixed     bool
	Source             PropertySource // which property sources to scan (default: all)
	IncludeValueCounts bool
}

type propertyCounts struct {
	noteCount   int
	shapes      map[string]int
	types       map[string]int
	values      map[string]struct{}
	valueCounts map[string]int
	overflow    bool
}

// Properties returns summaries for all frontmatter properties in the vault.
func Properties(vault obsidian.VaultManager, note obsidian.NoteManager, opts PropertiesOptions) ([]PropertySummary, error) {
	enumThreshold := opts.EnumThreshold
	if enumThreshold <= 0 {
		enumThreshold = 10
	}
	maxValues := opts.MaxValues
	if maxValues <= 0 {
		maxValues = 500
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	var scanNotes []string
	if len(opts.Notes) > 0 {
		scanNotes = opts.Notes
	} else {
		allNotes, err := note.GetNotesList(vaultPath)
		if err != nil {
			return nil, err
		}
		scanNotes = allNotes
	}

	numWorkers := runtime.NumCPU()
	if len(scanNotes) < numWorkers {
		numWorkers = len(scanNotes)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	batchSize := (len(scanNotes) + numWorkers - 1) / numWorkers
	results := make(chan map[string]*propertyCounts, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(scanNotes) {
			end = len(scanNotes)
		}
		if start >= len(scanNotes) {
			continue
		}

		batch := scanNotes[start:end]
		wg.Add(1)
		go func(files []string) {
			defer wg.Done()
			local := make(map[string]*propertyCounts)
			for _, notePath := range files {
				content, err := note.GetContents(vaultPath, notePath)
				if err != nil {
					continue
				}
				var frontmatter map[string]interface{}
				var inline map[string][]string
				if opts.Source != PropertySourceInline {
					frontmatter, _ = obsidian.ExtractFrontmatter(content)
				}
				if opts.Source != PropertySourceFrontmatter {
					inline = obsidian.ExtractInlineProperties(content)
				}

				perNote := make(map[string]*propertyCounts)

				addPerNote := func(key string, info obsidian.PropertyValueInfo) {
					if key == "tags" && opts.ExcludeTags {
						return
					}
					np, ok := perNote[key]
					if !ok {
						np = &propertyCounts{
							shapes:      make(map[string]int),
							types:       make(map[string]int),
							values:      make(map[string]struct{}),
							valueCounts: make(map[string]int),
						}
						perNote[key] = np
					}
					np.shapes[info.Shape]++
					np.types[info.ValueType]++
					for _, v := range info.Values {
						v = strings.TrimSpace(v)
						if v == "" {
							continue
						}
						if _, seen := np.values[v]; seen {
							continue
						}
						np.values[v] = struct{}{}
						np.valueCounts[v]++
					}
				}

				for key, raw := range frontmatter {
					info := obsidian.AnalyzePropertyValue(raw)
					addPerNote(key, info)
				}

				for key, values := range inline {
					info := obsidian.AnalyzePropertyValue(values)
					info.Shape = "scalar" // inline treated as scalar entries per line
					addPerNote(key, info)
				}

				for key, noteCounts := range perNote {
					pc, ok := local[key]
					if !ok {
						pc = &propertyCounts{
							shapes:      make(map[string]int),
							types:       make(map[string]int),
							values:      make(map[string]struct{}),
							valueCounts: make(map[string]int),
						}
						local[key] = pc
					}
					pc.noteCount++
					for s, c := range noteCounts.shapes {
						pc.shapes[s] += c
					}
					for t, c := range noteCounts.types {
						pc.types[t] += c
					}
					for v := range noteCounts.values {
						if key == "tags" {
							pc.values[v] = struct{}{}
							pc.valueCounts[v]++
							continue
						}
						if len(pc.values) < maxValues {
							pc.values[v] = struct{}{}
						} else {
							pc.overflow = true
							continue
						}
						pc.valueCounts[v]++
					}
				}
			}
			results <- local
		}(batch)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	merged := make(map[string]*propertyCounts)
	for res := range results {
		for key, pc := range res {
			target, ok := merged[key]
			if !ok {
				merged[key] = pc
				continue
			}
			target.noteCount += pc.noteCount
			for s, c := range pc.shapes {
				target.shapes[s] += c
			}
			for t, c := range pc.types {
				target.types[t] += c
			}
			for v := range pc.values {
				if key == "tags" {
					target.values[v] = struct{}{}
					continue
				}
				if len(target.values) < maxValues {
					target.values[v] = struct{}{}
				} else {
					target.overflow = true
					break
				}
			}
			if pc.overflow && key != "tags" {
				target.overflow = true
			}
			for v, c := range pc.valueCounts {
				target.valueCounts[v] += c
			}
		}
	}

	summaries := make([]PropertySummary, 0, len(merged))
	for key, pc := range merged {
		shape := pickOrMixed(pc.shapes)
		valueType := pickOrMixed(pc.types)

		var enumValues []string
		shouldEnumerate := (isEnumCandidate(valueType) || (valueType == "mixed" && len(pc.values) <= enumThreshold)) && len(pc.values) > 0 && len(pc.values) <= enumThreshold && !pc.overflow
		if key == "tags" && len(pc.values) > 0 {
			shouldEnumerate = true
		}

		if shouldEnumerate {
			enumValues = make([]string, 0, len(pc.values))
			for v := range pc.values {
				enumValues = append(enumValues, v)
			}
			sort.Strings(enumValues)
		}

		var enumValueCounts map[string]int
		if shouldEnumerate && opts.IncludeValueCounts && len(pc.valueCounts) > 0 {
			enumValueCounts = make(map[string]int, len(pc.valueCounts))
			for v, c := range pc.valueCounts {
				enumValueCounts[v] = c
			}
		}

		summaries = append(summaries, PropertySummary{
			Name:               key,
			NoteCount:          pc.noteCount,
			Shape:              shape,
			ValueType:          valueType,
			EnumValues:         enumValues,
			EnumValueCounts:    enumValueCounts,
			DistinctValueCount: len(pc.values),
			TruncatedValueSet:  pc.overflow,
		})
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].NoteCount == summaries[j].NoteCount {
			return summaries[i].Name < summaries[j].Name
		}
		return summaries[i].NoteCount > summaries[j].NoteCount
	})

	return summaries, nil
}

func pickOrMixed(counts map[string]int) string {
	if len(counts) == 0 {
		return "unknown"
	}
	if len(counts) == 1 {
		for k := range counts {
			return k
		}
	}
	return "mixed"
}

func isEnumCandidate(valueType string) bool {
	switch valueType {
	case "url", "datetime", "date", "object", "mixed", "unknown":
		return false
	default:
		return true
	}
}
