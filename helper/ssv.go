package helper

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/hdt3213/rdb/model"
)

func stringToSsvLine(obj *model.StringObject, fields []string) string {
	var ssvField = []string{strconv.Itoa(obj.DB)}
	if len(fields) > 0 {
		for _, i := range fields {
			switch i {
			case "key":
				ssvField = append(ssvField, obj.Key)
			case "expiration":
				if obj.Expiration != nil {
					ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
				}
			case "size":
				ssvField = append(ssvField, strconv.Itoa(obj.Size))
			case "type":
				ssvField = append(ssvField, obj.Type)
			case "encoding":
				ssvField = append(ssvField, obj.Encoding)
			case "value":
				if obj.Value != nil {
					ssvField = append(ssvField, strings.Replace(string(obj.Value), "\n", "", -1))
				}
			}

		}
	} else {
		ssvField = append(ssvField, obj.Key)
		if obj.Expiration != nil {
			ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
		}
		ssvField = append(ssvField, strconv.Itoa(obj.Size))
		ssvField = append(ssvField, obj.Type)
		ssvField = append(ssvField, obj.Encoding)
		if obj.Value != nil {
			ssvField = append(ssvField, strings.Replace(string(obj.Value), "\n", "", -1))
		}
	}

	return strings.Join(ssvField, " ")
}

func listToSsvLine(obj *model.ListObject, fields []string) (ssvLine string, err error) {
	var ssvField = []string{strconv.Itoa(obj.DB)}
	if len(fields) > 0 {
		for _, i := range fields {
			switch i {
			case "key":
				ssvField = append(ssvField, obj.Key)
			case "expiration":
				if obj.Expiration != nil {
					ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
				}
			case "size":
				ssvField = append(ssvField, strconv.Itoa(obj.Size))
			case "type":
				ssvField = append(ssvField, obj.Type)
			case "encoding":
				ssvField = append(ssvField, obj.Encoding)
			case "values":
				if obj.Values != nil {
					values, err := json.Marshal(obj.Values)
					if err != nil {
						return ssvLine, err
					}
					ssvField = append(ssvField, strings.Replace(string(values), "\n", "", -1))
				}
			}
		}
	} else {
		ssvField = append(ssvField, obj.Key)
		if obj.Expiration != nil {
			ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
		}
		ssvField = append(ssvField, strconv.Itoa(obj.Size))
		ssvField = append(ssvField, obj.Type)
		ssvField = append(ssvField, obj.Encoding)
		if obj.Values != nil {
			values, err := json.Marshal(obj.Values)
			if err != nil {
				return ssvLine, err
			}

			ssvField = append(ssvField, strings.Replace(string(values), "\n", "", -1))
		}
	}

	return strings.Join(ssvField, " "), nil
}

func hashToSsvLine(obj *model.HashObject, fields []string) (ssvLine string, err error) {
	var ssvField = []string{strconv.Itoa(obj.DB)}
	if len(fields) > 0 {
		for _, i := range fields {
			switch i {
			case "key":
				ssvField = append(ssvField, obj.Key)
			case "expiration":
				if obj.Expiration != nil {
					ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
				}
			case "size":
				ssvField = append(ssvField, strconv.Itoa(obj.Size))
			case "type":
				ssvField = append(ssvField, obj.Type)
			case "encoding":
				ssvField = append(ssvField, obj.Encoding)
			case "hash":
				if obj.Hash != nil {
					hash, err := json.Marshal(obj.Hash)
					if err != nil {
						return ssvLine, err
					}
					ssvField = append(ssvField, strings.Replace(string(hash), "\n", "", -1))
				}
			}

		}
	} else {
		ssvField = append(ssvField, obj.Key)
		if obj.Expiration != nil {
			ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
		}
		ssvField = append(ssvField, strconv.Itoa(obj.Size))
		ssvField = append(ssvField, obj.Type)
		ssvField = append(ssvField, obj.Encoding)
		if obj.Hash != nil {
			hash, err := json.Marshal(obj.Hash)
			if err != nil {
				return ssvLine, err
			}
			ssvField = append(ssvField, strings.Replace(string(hash), "\n", "", -1))
		}
	}

	return strings.Join(ssvField, " "), err
}

func setToSsvLine(obj *model.SetObject, fields []string) (ssvLine string, err error) {
	var ssvField = []string{strconv.Itoa(obj.DB)}
	if len(fields) > 0 {
		for _, i := range fields {
			switch i {
			case "key":
				ssvField = append(ssvField, obj.Key)
			case "expiration":
				if obj.Expiration != nil {
					ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
				}
			case "size":
				ssvField = append(ssvField, strconv.Itoa(obj.Size))
			case "type":
				ssvField = append(ssvField, obj.Type)
			case "encoding":
				ssvField = append(ssvField, obj.Encoding)
			case "members":
				if obj.Members != nil {
					members, err := json.Marshal(obj.Members)
					if err != nil {
						return ssvLine, err
					}
					ssvField = append(ssvField, strings.Replace(string(members), "\n", "", -1))
				}
			}

		}
	} else {
		ssvField = append(ssvField, obj.Key)
		if obj.Expiration != nil {
			ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
		}
		ssvField = append(ssvField, strconv.Itoa(obj.Size))
		ssvField = append(ssvField, obj.Type)
		ssvField = append(ssvField, obj.Encoding)
		if obj.Members != nil {
			members, err := json.Marshal(obj.Members)
			if err != nil {
				return ssvLine, err
			}
			ssvField = append(ssvField, strings.Replace(string(members), "\n", "", -1))
		}
	}

	return strings.Join(ssvField, " "), nil
}

func zSetToSsvLine(obj *model.ZSetObject, fields []string) (ssvLine string, err error) {
	var ssvField = []string{strconv.Itoa(obj.DB)}
	if len(fields) > 0 {
		for _, i := range fields {
			switch i {
			case "key":
				ssvField = append(ssvField, obj.Key)
			case "expiration":
				if obj.Expiration != nil {
					ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
				}
			case "size":
				ssvField = append(ssvField, strconv.Itoa(obj.Size))
			case "type":
				ssvField = append(ssvField, obj.Type)
			case "encoding":
				ssvField = append(ssvField, obj.Encoding)
			case "entries":
				if obj.Entries != nil {
					entries, err := json.Marshal(obj.Entries)
					if err != nil {
						return ssvLine, err
					}
					ssvField = append(ssvField, strings.Replace(string(entries), "\n", "", -1))
				}
			}

		}
	} else {
		ssvField = append(ssvField, obj.Key)
		if obj.Expiration != nil {
			ssvField = append(ssvField, obj.Expiration.Format("2006-01-02T15:04-07:00"))
		}
		ssvField = append(ssvField, strconv.Itoa(obj.Size))
		ssvField = append(ssvField, obj.Type)
		ssvField = append(ssvField, obj.Encoding)
		if obj.Entries != nil {
			entries, err := json.Marshal(obj.Entries)
			if err != nil {
				return ssvLine, err
			}
			ssvField = append(ssvField, strings.Replace(string(entries), "\n", "", -1))
		}
	}

	return strings.Join(ssvField, " "), nil
}

// ObjectToSsv convert redis object to space separated values
func ObjectToSsv(obj model.RedisObject, dec decoder) (ssvBytes []byte, err error) {
	if obj == nil {
		return ssvBytes, nil
	}

	ssvLines := make([]string, 0)
	switch obj.GetType() {
	case model.StringType:
		strObj := obj.(*model.StringObject)
		ssvLines = append(ssvLines, stringToSsvLine(strObj, dec.GetFields()))
	case model.ListType:
		listObj := obj.(*model.ListObject)
		line, err := listToSsvLine(listObj, dec.GetFields())
		if err != nil {
			return ssvBytes, err
		}
		ssvLines = append(ssvLines, line)
	case model.HashType:
		hashObj := obj.(*model.HashObject)
		line, err := hashToSsvLine(hashObj, dec.GetFields())
		if err != nil {
			return ssvBytes, err
		}
		ssvLines = append(ssvLines, line)
	case model.SetType:
		setObj := obj.(*model.SetObject)
		line, err := setToSsvLine(setObj, dec.GetFields())
		if err != nil {
			return ssvBytes, err
		}
		ssvLines = append(ssvLines, line)
	case model.ZSetType:
		zsetObj := obj.(*model.ZSetObject)
		line, err := zSetToSsvLine(zsetObj, dec.GetFields())
		if err != nil {
			return ssvBytes, err
		}
		ssvLines = append(ssvLines, line)
	}

	return []byte(strings.Join(ssvLines, "\n")), nil
}
