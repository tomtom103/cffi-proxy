package uagen

// import (
// 	"archive/zip"
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"math/rand"
// 	"os"
// 	"path/filepath"
// 	"time"
// )

// // BayesianNode represents a single node in a Bayesian network, allowing sampling from its conditional distribution.
// type BayesianNode struct {
// 	Name                     string
// 	ParentNames              []string
// 	PossibleValues           []string
// 	ConditionalProbabilities interface{}
// }

// // NewBayesianNode creates a new BayesianNode from a node definition.
// func NewBayesianNode(nodeDefinition map[string]interface{}) *BayesianNode {
// 	name := nodeDefinition["name"].(string)

// 	parentNames := []string{}
// 	if pn, ok := nodeDefinition["parentNames"].([]interface{}); ok {
// 		for _, p := range pn {
// 			parentNames = append(parentNames, p.(string))
// 		}
// 	}

// 	possibleValues := []string{}
// 	if pv, ok := nodeDefinition["possibleValues"].([]interface{}); ok {
// 		for _, v := range pv {
// 			possibleValues = append(possibleValues, v.(string))
// 		}
// 	}

// 	conditionalProbabilities := nodeDefinition["conditionalProbabilities"]

// 	return &BayesianNode{
// 		Name:                     name,
// 		ParentNames:              parentNames,
// 		PossibleValues:           possibleValues,
// 		ConditionalProbabilities: conditionalProbabilities,
// 	}
// }

// // GetProbabilitiesGivenKnownValues extracts unconditional probabilities of node values given the values of the parent nodes.
// func (bn *BayesianNode) GetProbabilitiesGivenKnownValues(parentValues map[string]interface{}) (map[string]float64, error) {
// 	probabilities := bn.ConditionalProbabilities

// 	for _, parentName := range bn.ParentNames {
// 		parentValue, ok := parentValues[parentName]
// 		if !ok {
// 			return nil, errors.New("parent value not found")
// 		}

// 		probabilitiesMap, ok := probabilities.(map[string]interface{})
// 		if !ok {
// 			return nil, errors.New("invalid probabilities structure")
// 		}

// 		deeper, hasDeeper := probabilitiesMap["deeper"]
// 		if hasDeeper {
// 			deeperMap, ok := deeper.(map[string]interface{})
// 			if !ok {
// 				return nil, errors.New("invalid deeper structure")
// 			}
// 			nextProbabilities, exists := deeperMap[parentValue.(string)]
// 			if exists {
// 				probabilities = nextProbabilities
// 			} else {
// 				skip, hasSkip := probabilitiesMap["skip"]
// 				if hasSkip {
// 					probabilities = skip
// 				} else {
// 					return nil, errors.New("no matching probabilities found")
// 				}
// 			}
// 		} else {
// 			skip, hasSkip := probabilitiesMap["skip"]
// 			if hasSkip {
// 				probabilities = skip
// 			} else {
// 				return nil, errors.New("no deeper or skip in probabilities")
// 			}
// 		}
// 	}

// 	// At this point, probabilities should be a map of values to probabilities
// 	result := make(map[string]float64)
// 	probabilitiesMap, ok := probabilities.(map[string]interface{})
// 	if !ok {
// 		return nil, errors.New("invalid probabilities at leaf node")
// 	}
// 	for k, v := range probabilitiesMap {
// 		if prob, ok := v.(float64); ok {
// 			result[k] = prob
// 		}
// 	}

// 	return result, nil
// }

// // SampleRandomValueFromPossibilities randomly samples from the given values using the given probabilities.
// func SampleRandomValueFromPossibilities(possibleValues []string, probabilities map[string]float64) (string, error) {
// 	rand.Seed(time.Now().UnixNano())
// 	anchor := rand.Float64()
// 	cumulativeProbability := 0.0

// 	for _, value := range possibleValues {
// 		prob, exists := probabilities[value]
// 		if !exists {
// 			return "", errors.New("probability not found for value")
// 		}
// 		cumulativeProbability += prob
// 		if cumulativeProbability > anchor {
// 			return value, nil
// 		}
// 	}
// 	// Default to the first item
// 	return possibleValues[0], nil
// }

// // Sample randomly samples from the conditional distribution of this node given values of parents.
// func (bn *BayesianNode) Sample(parentValues map[string]interface{}) (string, error) {
// 	probabilities, err := bn.GetProbabilitiesGivenKnownValues(parentValues)
// 	if err != nil {
// 		return "", err
// 	}
// 	value, err := SampleRandomValueFromPossibilities(bn.PossibleValues, probabilities)
// 	if err != nil {
// 		return "", err
// 	}
// 	return value, nil
// }

// // SampleAccordingToRestrictions randomly samples from the conditional distribution of this node given restrictions.
// func (bn *BayesianNode) SampleAccordingToRestrictions(parentValues map[string]interface{}, valuePossibilities []string, bannedValues []string) (string, error) {
// 	probabilities, err := bn.GetProbabilitiesGivenKnownValues(parentValues)
// 	if err != nil {
// 		return "", err
// 	}

// 	validValues := []string{}
// 	bannedSet := make(map[string]struct{})
// 	for _, val := range bannedValues {
// 		bannedSet[val] = struct{}{}
// 	}

// 	for _, val := range valuePossibilities {
// 		if _, banned := bannedSet[val]; !banned {
// 			if _, exists := probabilities[val]; exists {
// 				validValues = append(validValues, val)
// 			}
// 		}
// 	}

// 	if len(validValues) > 0 {
// 		value, err := SampleRandomValueFromPossibilities(validValues, probabilities)
// 		if err != nil {
// 			return "", err
// 		}
// 		return value, nil
// 	}
// 	return "", nil // Equivalent to None in Python
// }

// // BayesianNetwork represents a Bayesian network capable of randomly sampling from its distribution.
// type BayesianNetwork struct {
// 	NodesInSamplingOrder []*BayesianNode
// 	NodesByName          map[string]*BayesianNode
// }

// // NewBayesianNetwork creates a new BayesianNetwork from a JSON file path.
// func NewBayesianNetwork(path string) (*BayesianNetwork, error) {
// 	networkDefinition, err := ExtractJSON(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	nodesData, ok := networkDefinition["nodes"].([]interface{})
// 	if !ok {
// 		return nil, errors.New("invalid nodes data")
// 	}

// 	nodesInSamplingOrder := make([]*BayesianNode, len(nodesData))
// 	nodesByName := make(map[string]*BayesianNode)

// 	for i, nodeData := range nodesData {
// 		nodeDef, ok := nodeData.(map[string]interface{})
// 		if !ok {
// 			return nil, errors.New("invalid node definition")
// 		}
// 		node := NewBayesianNode(nodeDef)
// 		nodesInSamplingOrder[i] = node
// 		nodesByName[node.Name] = node
// 	}

// 	return &BayesianNetwork{
// 		NodesInSamplingOrder: nodesInSamplingOrder,
// 		NodesByName:          nodesByName,
// 	}, nil
// }

// // GenerateSample randomly samples from the distribution represented by the Bayesian network.
// func (bn *BayesianNetwork) GenerateSample(inputValues map[string]interface{}) (map[string]interface{}, error) {
// 	sample := make(map[string]interface{})
// 	for k, v := range inputValues {
// 		sample[k] = v
// 	}
// 	for _, node := range bn.NodesInSamplingOrder {
// 		if _, exists := sample[node.Name]; !exists {
// 			value, err := node.Sample(sample)
// 			if err != nil {
// 				return nil, err
// 			}
// 			sample[node.Name] = value
// 		}
// 	}
// 	return sample, nil
// }

// // GenerateConsistentSampleWhenPossible randomly samples values consistent with the provided restrictions.
// func (bn *BayesianNetwork) GenerateConsistentSampleWhenPossible(valuePossibilities map[string][]string) (map[string]interface{}, error) {
// 	return bn.recursivelyGenerateConsistentSampleWhenPossible(make(map[string]interface{}), valuePossibilities, 0)
// }

// func (bn *BayesianNetwork) recursivelyGenerateConsistentSampleWhenPossible(sampleSoFar map[string]interface{}, valuePossibilities map[string][]string, depth int) (map[string]interface{}, error) {
// 	if depth == len(bn.NodesInSamplingOrder) {
// 		return sampleSoFar, nil
// 	}

// 	node := bn.NodesInSamplingOrder[depth]
// 	bannedValues := []string{}
// 	for {
// 		possibilities, exists := valuePossibilities[node.Name]
// 		if !exists {
// 			possibilities = node.PossibleValues
// 		}
// 		value, err := node.SampleAccordingToRestrictions(sampleSoFar, possibilities, bannedValues)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if value == "" {
// 			break
// 		}
// 		sampleSoFar[node.Name] = value
// 		nextSample, err := bn.recursivelyGenerateConsistentSampleWhenPossible(sampleSoFar, valuePossibilities, depth+1)
// 		if err == nil {
// 			return nextSample, nil
// 		}
// 		bannedValues = append(bannedValues, value)
// 		delete(sampleSoFar, node.Name)
// 	}
// 	return nil, errors.New("no consistent sample found")
// }

// // Helper Functions

// // ArrayIntersection performs a set intersection on the given arrays.
// func ArrayIntersection(a, b []string) []string {
// 	setB := make(map[string]struct{})
// 	for _, val := range b {
// 		setB[val] = struct{}{}
// 	}
// 	intersection := []string{}
// 	for _, val := range a {
// 		if _, exists := setB[val]; exists {
// 			intersection = append(intersection, val)
// 		}
// 	}
// 	return intersection
// }

// // ArrayZip combines two arrays into a single array using the set union.
// func ArrayZip(a, b [][]string) [][]string {
// 	result := [][]string{}
// 	for i := range a {
// 		set := make(map[string]struct{})
// 		for _, val := range a[i] {
// 			set[val] = struct{}{}
// 		}
// 		for _, val := range b[i] {
// 			set[val] = struct{}{}
// 		}
// 		union := []string{}
// 		for val := range set {
// 			union = append(union, val)
// 		}
// 		result = append(result, union)
// 	}
// 	return result
// }

// // Undeeper removes the "deeper/skip" structures from the conditional probability table.
// func Undeeper(obj interface{}) interface{} {
// 	switch v := obj.(type) {
// 	case map[string]interface{}:
// 		result := make(map[string]interface{})
// 		for key, value := range v {
// 			if key == "skip" {
// 				continue
// 			}
// 			if key == "deeper" {
// 				deeperResult := Undeeper(value)
// 				if deeperMap, ok := deeperResult.(map[string]interface{}); ok {
// 					for k, val := range deeperMap {
// 						result[k] = val
// 					}
// 				}
// 			} else {
// 				result[key] = Undeeper(value)
// 			}
// 		}
// 		return result
// 	default:
// 		return obj
// 	}
// }

// // FilterByLastLevelKeys performs DFS on the tree and returns values of nodes on paths that end with the given keys.
// func FilterByLastLevelKeys(tree map[string]interface{}, validKeys []string) [][]string {
// 	var out [][]string

// 	var recurse func(t map[string]interface{}, vk []string, acc []string)
// 	recurse = func(t map[string]interface{}, vk []string, acc []string) {
// 		for key, value := range t {
// 			if _, isMap := value.(map[string]interface{}); !isMap || value == nil {
// 				for _, vkItem := range vk {
// 					if key == vkItem {
// 						if len(out) == 0 {
// 							for _, item := range acc {
// 								out = append(out, []string{item})
// 							}
// 						} else {
// 							for i := range out {
// 								out[i] = append(out[i], acc...)
// 							}
// 						}
// 					}
// 				}
// 				continue
// 			} else {
// 				recurse(value.(map[string]interface{}), vk, append(acc, key))
// 			}
// 		}
// 	}

// 	recurse(tree, validKeys, []string{})
// 	return out
// }

// // GetPossibleValues returns an extended set of constraints induced by the original constraints and network structure.
// func GetPossibleValues(network *BayesianNetwork, possibleValues map[string][]string) (map[string][]string, error) {
// 	sets := []map[string][]string{}

// 	for key, value := range possibleValues {
// 		if len(value) == 0 {
// 			return nil, errors.New("constraints are too restrictive")
// 		}
// 		node, exists := network.NodesByName[key]
// 		if !exists {
// 			return nil, errors.New("node not found in network")
// 		}
// 		tree := Undeeper(node.ConditionalProbabilities).(map[string]interface{})
// 		zippedValues := FilterByLastLevelKeys(tree, value)
// 		set := make(map[string][]string)
// 		for idx, parentName := range node.ParentNames {
// 			values := []string{}
// 			for _, zippedValue := range zippedValues {
// 				if idx < len(zippedValue) {
// 					values = append(values, zippedValue[idx])
// 				}
// 			}
// 			set[parentName] = values
// 		}
// 		set[key] = value
// 		sets = append(sets, set)
// 	}

// 	result := make(map[string][]string)
// 	for _, setDict := range sets {
// 		for key, values := range setDict {
// 			if existingValues, exists := result[key]; exists {
// 				intersectedValues := ArrayIntersection(values, existingValues)
// 				if len(intersectedValues) == 0 {
// 					return nil, errors.New("constraints are too restrictive")
// 				}
// 				result[key] = intersectedValues
// 			} else {
// 				result[key] = values
// 			}
// 		}
// 	}

// 	return result, nil
// }

// // ExtractJSON unzips a zip file if the path points to a zip file, otherwise directly loads a JSON file.
// func ExtractJSON(path string) (map[string]interface{}, error) {
// 	if filepath.Ext(path) != ".zip" {
// 		// Directly load the JSON file
// 		file, err := os.Open(path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer file.Close()
// 		return readJSON(file)
// 	}
// 	// Unzip the file and load the JSON content
// 	zipReader, err := zip.OpenReader(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer zipReader.Close()

// 	var jsonFile *zip.File
// 	for _, file := range zipReader.File {
// 		if filepath.Ext(file.Name) == ".json" {
// 			jsonFile = file
// 			break
// 		}
// 	}
// 	if jsonFile == nil {
// 		return nil, errors.New("no JSON file found in zip archive")
// 	}

// 	rc, err := jsonFile.Open()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rc.Close()

// 	return readJSON(rc)
// }

// func readJSON(r io.Reader) (map[string]interface{}, error) {
// 	var data map[string]interface{}
// 	decoder := json.NewDecoder(r)
// 	if err := decoder.Decode(&data); err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }
