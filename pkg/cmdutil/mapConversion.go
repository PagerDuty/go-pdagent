package cmdutil

func StringMapToInterfaceMap(stringMap map[string]string) map[string]interface{} {
	interfaceMap := map[string]interface{}{}
	for k, v := range stringMap {
		interfaceMap[k] = v
	}
	return interfaceMap
}
