package cmdutil

func ValidateMapFieldIsString(inputMap map[string]interface{}, selector string) (string, bool) {
	if value, ok := inputMap[selector]; ok {
		result, isResultString := value.(string)
		return result, isResultString
	}
	return "", false
}

func ValidateMapFieldIsMap(inputMap map[string]interface{}, selector string) (map[string]interface{}, bool) {
	if value, ok := inputMap[selector]; ok {
		result, isResultMap := value.(map[string]interface{})
		return result, isResultMap
	}
	return map[string]interface{}{}, false
}

func ValidateEnumField(inputVal string, allowedValues []string, err error) error {
	for _, value := range allowedValues {
		if value == inputVal {
			return nil
		}
	}
	return err
}
