package yqlib

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type unityYamlDecoder struct {
	decoder yaml.Decoder

	prefs YamlPreferences

	// work around of various parsing issues by yaml.v3 with document headers
	leadingContent string
	bufferRead     bytes.Buffer

	// anchor map persists over multiple documents for convenience.
	anchorMap map[string]*CandidateNode

	readAnything  bool
	firstFile     bool
	documentIndex uint
}

func NewUnityYamlDecoder(prefs YamlPreferences) Decoder {
	return &unityYamlDecoder{prefs: prefs, firstFile: true}
}

func (dec *unityYamlDecoder) processReadStream(reader *bufio.Reader) (io.Reader, string, error) {
	var commentLineRegEx = regexp.MustCompile(`^\s*#`)
	var yamlDirectiveLineRegEx = regexp.MustCompile(`^\s*%YA`)
	var tagDirectiveLineRegEx = regexp.MustCompile(`^\s*%TAG`)
	var sb strings.Builder
	var contentBuffer strings.Builder
	
	// Flag to track if we're processing Unity YAML
	isUnityYaml := false
	
	for {
		peekBytes, err := reader.Peek(4)
		if errors.Is(err, io.EOF) {
			// EOF are handled else where..
			return strings.NewReader(contentBuffer.String()), sb.String(), nil
		} else if err != nil {
			return strings.NewReader(contentBuffer.String()), sb.String(), err
		} else if string(peekBytes[0]) == "\n" {
			line, err := reader.ReadString('\n')
			sb.WriteString(line)
			contentBuffer.WriteString(line)
			if errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), nil
			} else if err != nil {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
		} else if tagDirectiveLineRegEx.MatchString(string(peekBytes)) {
			// Read the entire TAG directive line
			line, err := reader.ReadString('\n')
			sb.WriteString(line)
			
			// Check if this is Unity's tag directive
			if strings.Contains(line, "tag:unity3d.com,2011:") {
				isUnityYaml = true
				// Skip the TAG directive for Unity YAML files
				// The yaml.v3 parser doesn't handle custom tag prefixes well
				continue
			} else {
				contentBuffer.WriteString(line)
			}
			
			if errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), nil
			} else if err != nil {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
		} else if string(peekBytes) == "--- " {
			line, err := reader.ReadString(' ')
			if isUnityYaml {
				// For Unity YAML, we need to handle the document separator with tag references
				contentBuffer.WriteString("---")
				// Read the rest of the line which contains the Unity tag
				restOfLine, _ := reader.ReadString('\n')
				// Strip the Unity tag reference (e.g., !u!114 &-8676750429411634268)
				// and just keep the anchor if present
				if strings.Contains(restOfLine, "&") {
					parts := strings.Split(restOfLine, "&")
					if len(parts) > 1 {
						contentBuffer.WriteString(" &" + strings.TrimSpace(parts[1]))
					}
				}
				contentBuffer.WriteString("\n")
				sb.WriteString(line + restOfLine)
			} else {
				sb.WriteString("$yqDocSeparator$\n")
				contentBuffer.WriteString("$yqDocSeparator$\n")
			}
			if errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), nil
			} else if err != nil {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
		} else if string(peekBytes) == "---\n" {
			_, err := reader.ReadString('\n')
			sb.WriteString("$yqDocSeparator$\n")
			contentBuffer.WriteString("$yqDocSeparator$\n")
			if errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), nil
			} else if err != nil {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
		} else if commentLineRegEx.MatchString(string(peekBytes)) || yamlDirectiveLineRegEx.MatchString(string(peekBytes)) {
			line, err := reader.ReadString('\n')
			sb.WriteString(line)
			contentBuffer.WriteString(line)
			if errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), nil
			} else if err != nil {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
		} else {
			// Read the rest of the content
			restBytes, err := io.ReadAll(reader)
			if err != nil && !errors.Is(err, io.EOF) {
				return strings.NewReader(contentBuffer.String()), sb.String(), err
			}
			rest := string(restBytes)
			
			if isUnityYaml {
				// Remove Unity tag references throughout the document
				// Pattern: {fileID: <number>, guid: <guid>, type: <number>}
				unityRefPattern := regexp.MustCompile(`\{fileID:\s*-?\d+(?:,\s*guid:\s*[a-f0-9]+,\s*type:\s*\d+)?\}`)
				rest = unityRefPattern.ReplaceAllString(rest, "null")
				
				// Remove inline Unity tags like !u!1
				unityTagPattern := regexp.MustCompile(`!u!\d+`)
				rest = unityTagPattern.ReplaceAllString(rest, "")
			}
			
			contentBuffer.WriteString(rest)
			return strings.NewReader(contentBuffer.String()), sb.String(), nil
		}
	}
}

func (dec *unityYamlDecoder) Init(reader io.Reader) error {
	readerToUse := reader
	leadingContent := ""
	dec.bufferRead = bytes.Buffer{}
	var err error
	// Always process Unity YAML files
	readerToUse, leadingContent, err = dec.processReadStream(bufio.NewReader(reader))
	if err != nil {
		return err
	}
	
	dec.leadingContent = leadingContent
	dec.readAnything = false
	dec.decoder = *yaml.NewDecoder(readerToUse)
	dec.firstFile = false
	dec.documentIndex = 0
	dec.anchorMap = make(map[string]*CandidateNode)
	return nil
}

func (dec *unityYamlDecoder) Decode() (*CandidateNode, error) {
	var yamlNode yaml.Node
	err := dec.decoder.Decode(&yamlNode)

	if errors.Is(err, io.EOF) && dec.leadingContent != "" && !dec.readAnything {
		// force returning an empty node with a comment.
		dec.readAnything = true
		return dec.blankNodeWithComment(), nil
	} else if errors.Is(err, io.EOF) && !dec.prefs.LeadingContentPreProcessing && !dec.readAnything {
		// didn't find any yaml,
		// check the tee buffer, maybe there were comments
		dec.readAnything = true
		dec.leadingContent = dec.bufferRead.String()
		if dec.leadingContent != "" {
			return dec.blankNodeWithComment(), nil
		}
		return nil, err
	} else if err != nil {
		return nil, err
	}

	candidateNode := CandidateNode{document: dec.documentIndex}
	// don't bother with the DocumentNode
	err = candidateNode.UnmarshalYAML(yamlNode.Content[0], dec.anchorMap)
	if err != nil {
		return nil, err
	}

	candidateNode.HeadComment = yamlNode.HeadComment + candidateNode.HeadComment
	candidateNode.FootComment = yamlNode.FootComment + candidateNode.FootComment

	if dec.leadingContent != "" {
		candidateNode.LeadingContent = dec.leadingContent
		dec.leadingContent = ""
	}
	dec.readAnything = true
	dec.documentIndex++
	return &candidateNode, nil
}

func (dec *unityYamlDecoder) blankNodeWithComment() *CandidateNode {
	node := createScalarNode(nil, "")
	node.LeadingContent = dec.leadingContent
	return node
}