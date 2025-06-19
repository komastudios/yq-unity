package yqlib

import (
	"fmt"
	"regexp"
	"strings"
)

// UnityGraphNode represents a node in a Unity graph asset
type UnityGraphNode struct {
	FileID          string
	Name            string
	ScriptType      string
	Properties      map[string]interface{}
	SerializedData  string
}

// ParseUnityGraphNodes extracts all nodes from a Unity graph asset
func ParseUnityGraphNodes(content string) ([]UnityGraphNode, error) {
	var nodes []UnityGraphNode
	
	// Split by document separator
	documents := strings.Split(content, "---")
	
	for _, doc := range documents {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		
		// Extract FileID from the document header
		fileIDPattern := regexp.MustCompile(`!u!\d+\s*&(-?\d+)`)
		fileIDMatches := fileIDPattern.FindStringSubmatch(doc)
		
		var fileID string
		if len(fileIDMatches) > 1 {
			fileID = fileIDMatches[1]
		}
		
		// Extract m_Name
		namePattern := regexp.MustCompile(`m_Name:\s*(.+)`)
		nameMatches := namePattern.FindStringSubmatch(doc)
		
		var name string
		if len(nameMatches) > 1 {
			name = strings.TrimSpace(nameMatches[1])
		}
		
		// Extract script type from m_Script
		scriptPattern := regexp.MustCompile(`m_Script:\s*\{fileID:\s*\d+,\s*guid:\s*([a-f0-9]+)`)
		scriptMatches := scriptPattern.FindStringSubmatch(doc)
		
		var scriptType string
		if len(scriptMatches) > 1 {
			scriptType = scriptMatches[1]
		}
		
		// Extract properties
		properties := make(map[string]interface{})
		
		// Common spawner properties
		propertyPatterns := map[string]*regexp.Regexp{
			"Extents":         regexp.MustCompile(`Extents:\s*(.+)`),
			"GridCellSize":    regexp.MustCompile(`GridCellSize:\s*(.+)`),
			"MinimumDistance": regexp.MustCompile(`MinimumDistance:\s*(.+)`),
			"RoadPoseDistance": regexp.MustCompile(`RoadPoseDistance:\s*(.+)`),
			"Spacing":         regexp.MustCompile(`Spacing:\s*(.+)`),
			"Jittering":       regexp.MustCompile(`Jittering:\s*(.+)`),
			"EligibleForInjection": regexp.MustCompile(`EligibleForInjection:\s*(.+)`),
		}
		
		for propName, pattern := range propertyPatterns {
			matches := pattern.FindStringSubmatch(doc)
			if len(matches) > 1 {
				properties[propName] = strings.TrimSpace(matches[1])
			}
		}
		
		// Extract serializedData if present
		serializedPattern := regexp.MustCompile(`serializedData:\s*'([^']*)'`)
		serializedMatches := serializedPattern.FindStringSubmatch(doc)
		
		var serializedData string
		if len(serializedMatches) > 1 {
			serializedData = serializedMatches[1]
		}
		
		if name != "" {
			nodes = append(nodes, UnityGraphNode{
				FileID:         fileID,
				Name:           name,
				ScriptType:     scriptType,
				Properties:     properties,
				SerializedData: serializedData,
			})
		}
	}
	
	return nodes, nil
}

// FilterNodesByType filters nodes by their name pattern
func FilterNodesByType(nodes []UnityGraphNode, typePattern string) []UnityGraphNode {
	var filtered []UnityGraphNode
	pattern := regexp.MustCompile(typePattern)
	
	for _, node := range nodes {
		if pattern.MatchString(node.Name) {
			filtered = append(filtered, node)
		}
	}
	
	return filtered
}

// ExtractSpawnerDistances extracts all distance-related properties from spawner nodes
func ExtractSpawnerDistances(nodes []UnityGraphNode) map[string]map[string]interface{} {
	spawnerDistances := make(map[string]map[string]interface{})
	
	spawnerNodes := FilterNodesByType(nodes, "(?i)(spawner|pose.*set)")
	
	for _, node := range spawnerNodes {
		if len(node.Properties) > 0 {
			// Filter only distance-related properties
			distanceProps := make(map[string]interface{})
			for key, value := range node.Properties {
				if strings.Contains(strings.ToLower(key), "distance") ||
				   strings.Contains(strings.ToLower(key), "extent") ||
				   strings.Contains(strings.ToLower(key), "spacing") ||
				   strings.Contains(strings.ToLower(key), "cell") ||
				   strings.Contains(strings.ToLower(key), "radius") {
					distanceProps[key] = value
				}
			}
			
			if len(distanceProps) > 0 {
				spawnerDistances[fmt.Sprintf("%s (%s)", node.Name, node.FileID)] = distanceProps
			}
		}
	}
	
	return spawnerDistances
}

// GetNodeByFileID finds a node by its FileID
func GetNodeByFileID(nodes []UnityGraphNode, fileID string) *UnityGraphNode {
	for _, node := range nodes {
		if node.FileID == fileID {
			return &node
		}
	}
	return nil
}

// GetNodeConnections extracts connections between nodes
func GetNodeConnections(content string, node UnityGraphNode) []string {
	var connections []string
	
	// Look for references to this node's FileID
	pattern := regexp.MustCompile(fmt.Sprintf(`\{fileID:\s*%s\}`, node.FileID))
	matches := pattern.FindAllStringIndex(content, -1)
	
	// For each match, find the containing document
	for _, match := range matches {
		// Find the document this reference is in
		beforeMatch := content[:match[0]]
		lastDocStart := strings.LastIndex(beforeMatch, "---")
		
		if lastDocStart >= 0 {
			nextDocStart := strings.Index(content[lastDocStart+3:], "---")
			var docContent string
			if nextDocStart >= 0 {
				docContent = content[lastDocStart : lastDocStart+3+nextDocStart]
			} else {
				docContent = content[lastDocStart:]
			}
			
			// Extract the FileID of the referencing document
			fileIDPattern := regexp.MustCompile(`!u!\d+\s*&(-?\d+)`)
			fileIDMatches := fileIDPattern.FindStringSubmatch(docContent)
			if len(fileIDMatches) > 1 {
				connections = append(connections, fileIDMatches[1])
			}
		}
	}
	
	return connections
}