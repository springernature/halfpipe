package helpers

import "strings"

func SecretToMapAndKey(secret string) (string, string) {
	s := strings.Replace(strings.Replace(secret, "((", "", -1), "))", "", -1)
	parts := strings.Split(s, ".")
	mapName, keyName := parts[0], parts[1]
	return mapName, keyName
}

