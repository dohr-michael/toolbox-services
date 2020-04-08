package lib

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
	"time"
)

func mapStructureToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

func mapStructureDecode(input interface{}, result interface{}, tagName string) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: tagName,
		Squash:  true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapStructureToTimeHookFunc(),
		),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}
	return err
}

func FromMap(input map[string]interface{}, result interface{}, tagName string) error {
	return mapStructureDecode(input, result, tagName)
}

func AsMap(input interface{}, tagName string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := mapStructureDecode(input, &result, tagName)
	if err != nil {
		return nil, err
	}
	return result, nil
}
