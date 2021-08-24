package cmdutil

func GetNestedStringField(input map[string]interface{}, selectors ...string) (string, bool) {
	node := input
	for index, key := range selectors {
		if nextNode, ok := node[key]; ok {
			if index+1 == len(selectors) {
				result, ok := nextNode.(string)
				return result, ok
			} else if mapVal, ok := nextNode.(map[string]interface{}); ok {
				node = mapVal
			} else {
				break
			}
		} else {
			break
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
