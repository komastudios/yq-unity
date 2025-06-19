package yqlib

import (
	"container/list"
	"fmt"
	"os"
)

func unityGraphNodesOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	var results = list.New()

	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		candidate := el.Value.(*CandidateNode)
		
		// Read the entire file if we're at the document root
		content := ""
		if candidate.filename != "" {
			fileBytes, err := os.ReadFile(candidate.filename)
			if err != nil {
				return Context{}, fmt.Errorf("failed to read unity file: %w", err)
			}
			content = string(fileBytes)
		} else {
			// Fallback to node value
			if candidate.Kind == ScalarNode {
				content = candidate.Value
			}
		}
		
		// Parse Unity graph nodes
		nodes, err := ParseUnityGraphNodes(content)
		if err != nil {
			return Context{}, err
		}
		
		// Create a new map node with the parsed data
		mapNode := &CandidateNode{
			Kind: MappingNode,
			Tag:  "!!map",
			Content: make([]*CandidateNode, 0),
		}
		
		for _, node := range nodes {
			// Create node entry
			keyNode := createScalarNode(node.Name, "!!str")
			
			// Create properties map
			propsNode := &CandidateNode{
				Kind: MappingNode,
				Tag:  "!!map",
				Content: make([]*CandidateNode, 0),
			}
			
			// Add FileID
			if node.FileID != "" {
				propsNode.Content = append(propsNode.Content,
					createScalarNode("FileID", "!!str"),
					createScalarNode(node.FileID, "!!str"))
			}
			
			// Add properties
			for key, value := range node.Properties {
				propsNode.Content = append(propsNode.Content,
					createScalarNode(key, "!!str"),
					createScalarNode(fmt.Sprintf("%v", value), "!!str"))
			}
			
			mapNode.Content = append(mapNode.Content, keyNode, propsNode)
		}
		
		results.PushBack(mapNode)
	}

	return context.ChildContext(results), nil
}

func unitySpawnerDistancesOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	var results = list.New()

	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		candidate := el.Value.(*CandidateNode)
		
		// Read the entire file if we're at the document root
		content := ""
		if candidate.filename != "" {
			fileBytes, err := os.ReadFile(candidate.filename)
			if err != nil {
				return Context{}, fmt.Errorf("failed to read unity file: %w", err)
			}
			content = string(fileBytes)
		} else {
			// Fallback to node value
			if candidate.Kind == ScalarNode {
				content = candidate.Value
			}
		}
		
		// Parse Unity graph nodes
		nodes, err := ParseUnityGraphNodes(content)
		if err != nil {
			return Context{}, err
		}
		
		// Extract spawner distances
		spawnerDistances := ExtractSpawnerDistances(nodes)
		
		// Create a new map node with the spawner distances
		mapNode := &CandidateNode{
			Kind: MappingNode,
			Tag:  "!!map",
			Content: make([]*CandidateNode, 0),
		}
		
		for spawnerName, distances := range spawnerDistances {
			keyNode := createScalarNode(spawnerName, "!!str")
			
			// Create distances map
			distNode := &CandidateNode{
				Kind: MappingNode,
				Tag:  "!!map",
				Content: make([]*CandidateNode, 0),
			}
			
			for propName, propValue := range distances {
				distNode.Content = append(distNode.Content,
					createScalarNode(propName, "!!str"),
					createScalarNode(fmt.Sprintf("%v", propValue), "!!str"))
			}
			
			mapNode.Content = append(mapNode.Content, keyNode, distNode)
		}
		
		results.PushBack(mapNode)
	}

	return context.ChildContext(results), nil
}

