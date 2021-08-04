package cmdutil

import "strings"

func GetNestedStringField(inputMap map[string]interface{}, selector string) (string, bool) {
	selectors := strings.Split(selector, ".")
	currentMap := inputMap
	for index, key := range selectors {
		if value, ok := currentMap[key]; ok {
			if index+1 == len(selectors) {
				result, isResultString := value.(string)
				return result, isResultString
			} else if mapVal, isMapVal := value.(map[string]interface{}); isMapVal {
				currentMap = mapVal
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}
	return "", false
}

func StringMapToInterfaceMap(stringMap map[string]string) map[string]interface{} {
	interfaceMap := map[string]interface{}{}
	for k, v := range stringMap {
		interfaceMap[k] = v
	}
	return interfaceMap
}
