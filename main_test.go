package main

import "testing"

func TestExtractDuplicate(t *testing.T) {
  nameMap := make(map[string]bool)
  metaName := "P1020323.JPG(1).json"
  nameMap[metaName] = true
  expectedName := "P1020323(1).JPG"
  nameMap[expectedName] = true
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
  nameMap := make(map[string]bool)
  metaName := "IMG_1425-ANIMATION.gif(1).json"
  nameMap[metaName] = true
  expectedName := "IMG_1425-ANIMATION(1).gif"
  nameMap[expectedName] = true
  check(t, metaName, nameMap, expectedName)
}

func TestExtractSimple(t *testing.T) {
  nameMap := make(map[string]bool)
  metaName := "IMG_20180911_155532.jpg.json"
  nameMap[metaName] = true
  expectedName := "IMG_20180911_155532.jpg"
  nameMap[expectedName] = true
  check(t, metaName, nameMap, expectedName)
}

func TestExtractLong(t *testing.T) {
  nameMap := make(map[string]bool)
  metaName := "70ECD1A6-F846-4CB7-9709-474FCB7B3E15-COLLAGE.j.json"
  nameMap[metaName] = true
  expectedName := "70ECD1A6-F846-4CB7-9709-474FCB7B3E15-COLLAGE.jpg"
  nameMap[expectedName] = true
  check(t, metaName, nameMap, expectedName)
}

func check(t *testing.T, metaName string, nameMap map[string]bool, expectedName string) {
  name, err := getItemFileName(metaName, nameMap)
  if err != nil {
    t.Error(err)
    return
  }

  if name != expectedName {
    t.Errorf("%s != %s", name, expectedName)
  }
}
