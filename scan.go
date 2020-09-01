package main

import (
  "errors"
  "fmt"
  "github.com/valyala/fastjson"
  "io/ioutil"
  "log"
  "path/filepath"
  "strconv"
  "strings"
  "time"
)

var itemExtensions = []string{"jpg", "gif", "jpeg", "webm", "webp", "mp4", "mov", "mkv"}

func getItemFileName(metaName string, nameMap map[string]string) (string, error) {
  candidate := strings.ToLower(metaName[:len(metaName)-len(".json")])
  itemName := nameMap[candidate]
  if len(itemName) != 0 {
    return itemName, nil
  }

  // for creations with long name jpg is shortened to j
  if strings.HasSuffix(candidate, ".j") {
    itemName = nameMap[candidate+"pg"]
    if len(itemName) != 0 {
      return itemName, nil
    }
  }

  nameWithoutExtension := strings.ToLower(metaName[:len(metaName)-len("json")])
  for _, ext := range itemExtensions {
    // P1020323.JPG(1). -> P1020323(1).
    candidate = strings.ToLower(strings.Replace(nameWithoutExtension, "."+ext+"(", "(", 1)) + ext
    itemName = nameMap[candidate]
    if len(itemName) != 0 {
      return itemName, nil
    }
  }

  return "", fmt.Errorf("cannot determinate image file metaName (metaName=%s)", metaName)
}

func processMetaItem(
  metaName string,
  path string,
  dir string,
  nameMap map[string]string,
  jsonParser *fastjson.Parser,
  yearMap map[int]map[time.Month]map[ItemKey]ItemValue,
  years *[]int,
) error {
  metaBytes, err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }

  meta, err := jsonParser.ParseBytes(metaBytes)
  if err != nil {
    return err
  }

  if meta.Exists("albumData") {
    return nil
  }

  s := string(meta.GetStringBytes("photoTakenTime", "timestamp"))
  if len(s) == 0 {
    return errors.New("photoTakenTime is empty")
  }
  timestampRaw, err := strconv.ParseInt(s, 0, 64)
  if err != nil {
    return err
  }

  timestamp := time.Unix(timestampRaw, 0).UTC()
  monthMap, ok := yearMap[timestamp.Year()]
  if !ok {
    *years = append(*years, timestamp.Year())
    monthMap = make(map[time.Month]map[ItemKey]ItemValue)
    yearMap[timestamp.Year()] = monthMap
  }

  timeMap, ok := monthMap[timestamp.Month()]
  if !ok {
    timeMap = make(map[ItemKey]ItemValue)
    monthMap[timestamp.Month()] = timeMap
  }

  itemKey := ItemKey{
    metaName: metaName,
    time:     timestamp,
  }

  if _, exists := timeMap[itemKey]; exists {
    return nil
  }

  // title cannot be used as filename - it is title and in case of duplicated file name it still contain original name,
  // so, determinate by metafile name
  itemFileName, err := getItemFileName(metaName, nameMap)
  if err != nil {
    log.Printf("WARN: cannot find original file for %s\n", path)
    return nil
  }

  if len(itemFileName) == 0 {
    return errors.New("itemFileName is empty")
  }

  ext := filepath.Ext(itemFileName)
  edited := nameMap[strings.ToLower(itemFileName[:len(itemFileName)-len(ext)] + "-edited" + ext)]

  timeMap[itemKey] = ItemValue{
    dir:        dir,
    name:       itemFileName,
    editedName: edited,
  }
  return nil
}
