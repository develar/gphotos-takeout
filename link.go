package main

import (
  "bytes"
  "fmt"
  "log"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "time"
)

func linkItems(years []int, yearMap map[int]map[time.Month]map[ItemKey]ItemValue, outDir string) error {
  err := os.Mkdir(outDir, 0755)
  if err != nil && !os.IsExist(err) {
    return err
  }

  for _, year := range years {
    yearDir := filepath.Join(outDir, strconv.Itoa(year))
    err = os.Mkdir(yearDir, 0755)
    if err != nil && !os.IsExist(err) {
      return err
    }

    for month := 1; month < 13; month++ {
      itemMap, ok := yearMap[year][time.Month(month)]
      if !ok {
        continue
      }

      monthDir := filepath.Join(yearDir, fmt.Sprintf("%02d", month))
      err = os.Mkdir(monthDir, 0755)
      if err != nil && !os.IsExist(err) {
        return err
      }

      keys := getKeys(itemMap)
      writtenNames := make(map[string]ItemValue)
      for _, key := range keys {
        err = linkItem(itemMap[key], key, monthDir, writtenNames)
        if err != nil {
          return err
        }
      }
    }
  }
  return nil
}

func linkItem(item ItemValue, key ItemKey, monthDir string, writtenNames map[string]ItemValue) error {
  inImageFile := filepath.Join(item.dir, item.name)
  outFileName := strings.ToLower(item.name)
  outImageFile := filepath.Join(monthDir, outFileName)

  err := os.Link(inImageFile, outImageFile)

  timeSuffix := ""
  if err != nil {
    if !os.IsExist(err) {
      return err
    }

    // google takeout can duplicate files (album vs auto-upload) (different timestamps but the same file)
    aHash, err := hashFile(inImageFile)
    if err != nil {
      return err
    }

    bHash, err := hashFile(outImageFile)
    if err != nil {
      return err
    }

    if bytes.Equal(aHash, bHash) {
      // skip duplicated file
      return nil
    }

    timeSuffix = key.time.Format("02-150405")
    if !strings.EqualFold(outFileName, "color_pop.jpg") &&
      !strings.EqualFold(outFileName, "effects.jpg") &&
      !strings.EqualFold(outFileName, "movie.mp4") &&
      !strings.EqualFold(outFileName, "movie(1).mp4") {
      log.Printf("file name is duplicated but content is different, written with suffix %s (previousFile=%s, currentFile=%s, meta=%s)",
        timeSuffix,
        filepath.Join(writtenNames[item.name].dir, item.name),
        filepath.Join(item.dir, item.name),
        filepath.Join(item.dir, key.metaName),
      )
    }

    outFileName = nameWithSuffix(outFileName, timeSuffix)
    outImageFile := filepath.Join(monthDir, outFileName)
    err = os.Link(inImageFile, outImageFile)
    if err != nil {
      return err
    }
  }

  // ensure that meta file name is consistent - do not use google take-out original name
  err = os.Link(filepath.Join(item.dir, key.metaName), filepath.Join(monthDir, outFileName+".json"))
  if err != nil {
    return err
  }

  if len(item.editedName) != 0 {
    err = os.Link(filepath.Join(item.dir, item.editedName), filepath.Join(monthDir, nameWithSuffix(outFileName, "edited")))
    if err != nil {
      return err
    }
  }

  writtenNames[item.name+timeSuffix] = item
  return nil
}

func nameWithSuffix(s string, timeSuffix string) string {
  if len(timeSuffix) == 0 {
    return s
  }

  ext := filepath.Ext(s)
  return s[:len(s)-len(ext)] + "_" + timeSuffix + ext
}
