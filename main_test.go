package main

import (
  "strings"
  "testing"
)

func TestExtractDuplicate(t *testing.T) {
  nameMap := make(map[string]string)
  metaName := "P1020323.JPG(1).json"
  nameMap[strings.ToLower(metaName)] = metaName
  expectedName := "P1020323(1).JPG"
  nameMap[strings.ToLower(expectedName)] = expectedName
  name, err := getItemFileName(metaName, nameMap)
  if err != nil {
    t.Error(err)
    return
  }

  if name != expectedName {
    t.Errorf("%s != %s", name, expectedName)
  }
}

func TestExtractDuplicate2(t *testing.T) {
  nameMap := make(map[string]string)
  metaName := "IMG_1425-ANIMATION.gif(1).json"
  nameMap[strings.ToLower(metaName)] = metaName
  expectedName := "IMG_1425-ANIMATION(1).gif"
  nameMap[strings.ToLower(expectedName)] = expectedName
  check(t, metaName, nameMap, expectedName)
}

func TestExtractSimple(t *testing.T) {
  nameMap := make(map[string]string)
  metaName := "IMG_20180911_155532.jpg.json"
  nameMap[strings.ToLower(metaName)] = metaName
  expectedName := "IMG_20180911_155532.jpg"
  nameMap[strings.ToLower(expectedName)] = expectedName
  check(t, metaName, nameMap, expectedName)
}

func TestExtractLong(t *testing.T) {
  nameMap := make(map[string]string)
  metaName := "70ECD1A6-F846-4CB7-9709-474FCB7B3E15-COLLAGE.j.json"
  nameMap[strings.ToLower(metaName)] = metaName
  expectedName := "70ECD1A6-F846-4CB7-9709-474FCB7B3E15-COLLAGE.jpg"
  nameMap[strings.ToLower(expectedName)] = expectedName
  check(t, metaName, nameMap, expectedName)
}

func check(t *testing.T, metaName string, nameMap map[string]string, expectedName string) {
  name, err := getItemFileName(metaName, nameMap)
  if err != nil {
    t.Error(err)
    return
  }

  if name != expectedName {
    t.Errorf("%s != %s", name, expectedName)
  }
}
