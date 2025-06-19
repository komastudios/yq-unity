package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var unityExtractCmd = &cobra.Command{
	Use:   "unity-extract [spawners|nodes|property] file.asset",
	Short: "Extract data from Unity asset files",
	Long: `Extract specific data from Unity asset files.
	
Available extraction types:
  spawners - Extract all spawner nodes with distance properties
  nodes    - Extract all nodes with their properties
  property - Extract a specific property (requires --property flag)`,
	Args: cobra.MinimumNArgs(2),
	RunE: unityExtractFunc,
}

var propertyName string

func init() {
	unityExtractCmd.Flags().StringVar(&propertyName, "property", "", "Property name to extract")
}

func unityExtractFunc(cmd *cobra.Command, args []string) error {
	extractType := args[0]
	filename := args[1]

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	switch extractType {
	case "spawners":
		return extractSpawners(string(content))
	case "nodes":
		return extractNodes(string(content))
	case "property":
		if propertyName == "" {
			return fmt.Errorf("property extraction requires --property flag")
		}
		return extractProperty(string(content), propertyName)
	default:
		return fmt.Errorf("unknown extraction type: %s", extractType)
	}
}

func extractSpawners(content string) error {
	// Split by document separator
	documents := strings.Split(content, "---")
	
	spawnerCount := 0
	for _, doc := range documents {
		if strings.Contains(doc, "Spawner") && strings.Contains(doc, "m_Name:") {
			// Extract node name
			nameMatch := regexp.MustCompile(`m_Name:\s*(.+)`).FindStringSubmatch(doc)
			if len(nameMatch) > 1 {
				name := strings.TrimSpace(nameMatch[1])
				
				// Look for distance properties
				properties := make(map[string]string)
				
				// Common spawner properties
				propertyPatterns := map[string]*regexp.Regexp{
					"Extents":         regexp.MustCompile(`Extents:\s*(.+)`),
					"GridCellSize":    regexp.MustCompile(`GridCellSize:\s*(.+)`),
					"MinimumDistance": regexp.MustCompile(`MinimumDistance:\s*(.+)`),
					"RoadPoseDistance": regexp.MustCompile(`RoadPoseDistance:\s*(.+)`),
					"Spacing":         regexp.MustCompile(`Spacing:\s*(.+)`),
				}
				
				for propName, pattern := range propertyPatterns {
					matches := pattern.FindStringSubmatch(doc)
					if len(matches) > 1 {
						properties[propName] = strings.TrimSpace(matches[1])
					}
				}
				
				if len(properties) > 0 {
					spawnerCount++
					fmt.Printf("%s:\n", name)
					for k, v := range properties {
						fmt.Printf("  %s: %s\n", k, v)
					}
					fmt.Println()
				}
			}
		}
	}
	
	if spawnerCount == 0 {
		fmt.Println("No spawners with distance properties found")
	} else {
		fmt.Printf("Total spawners found: %d\n", spawnerCount)
	}
	
	return nil
}

func extractNodes(content string) error {
	// Split by document separator
	documents := strings.Split(content, "---")
	
	nodeCount := 0
	for _, doc := range documents {
		if strings.Contains(doc, "m_Name:") {
			// Extract node name
			nameMatch := regexp.MustCompile(`m_Name:\s*(.+)`).FindStringSubmatch(doc)
			if len(nameMatch) > 1 {
				name := strings.TrimSpace(nameMatch[1])
				nodeCount++
				fmt.Printf("- %s\n", name)
			}
		}
	}
	
	fmt.Printf("\nTotal nodes: %d\n", nodeCount)
	return nil
}

func extractProperty(content string, property string) error {
	// Create regex for the property
	pattern := regexp.MustCompile(fmt.Sprintf(`%s:\s*(.+)`, regexp.QuoteMeta(property)))
	
	matches := pattern.FindAllStringSubmatch(content, -1)
	
	if len(matches) == 0 {
		fmt.Printf("Property '%s' not found\n", property)
		return nil
	}
	
	// Count unique values
	values := make(map[string]int)
	for _, match := range matches {
		if len(match) > 1 {
			value := strings.TrimSpace(match[1])
			values[value]++
		}
	}
	
	fmt.Printf("Property '%s' values:\n", property)
	for value, count := range values {
		fmt.Printf("  %s: %d occurrences\n", value, count)
	}
	
	return nil
}