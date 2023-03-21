package map_type

func ToTypedObject(input map[string]interface{}) map[string]string {
	output := make(map[string]string)

	for k, v := range input {
		if v == nil {
			continue
		}

		output[k] = v.(string)
	}

	return output
}
