package main

import (
  "bytes"
  "log"
  "os"
  "path/filepath"
  "time"
)

// google takeout can duplicate files - VID_20200626_124037.mp4 and VID_20200626_124037(1).mp4
// same content and same photoTakenTime, but different creationTime
func deduplicate(years []int, yearMap map[int]map[time.Month]map[ItemKey]ItemValue, rootDir string) error {
  for _, year := range years {
    for month := 1; month < 13; month++ {
      itemMap, ok := yearMap[year][time.Month(month)]
      if !ok {
        continue
      }

      keys := getKeys(itemMap)
      timeMap := make(map[time.Time]ItemKey)

      for _, key := range keys {
        existingKey, exists := timeMap[key.time]
        if !exists {
          timeMap[key.time] = key
          continue
        }

        // same time - compare check content
        aPath := getFilePath(itemMap[key])
        bPath := getFilePath(itemMap[existingKey])
        equal, err := filesEqual(aPath, bPath)
        if err != nil {
          return err
        }
        if equal {
          delete(itemMap, key)
          log.Print(getRelativePath(rootDir, aPath) + " duplicates " + getRelativePath(rootDir, bPath))
        }
      }
    }
  }

  return nil
}

func getRelativePath(rootDir string, file string) string {
  result, err := filepath.Rel(rootDir, file)
  if err != nil {
    return file
  }
  return result
}

func filesEqual(aPath string, bPath string) (bool, error) {
  aStat, err := os.Stat(aPath)
  if err != nil {
    return false, err
  }

  bStat, err := os.Stat(bPath)
  if err != nil {
    return false, err
  }

  if aStat.Size() != bStat.Size() {
    // not a duplicate
    return false, nil
  }

  aHash, err := hashFile(aPath)
  if err != nil {
    return false, err
  }

  bHash, err := hashFile(bPath)
  if err != nil {
    return false, err
  }

  return bytes.Equal(aHash, bHash), nil
}

func getFilePath(value ItemValue) string {
  return filepath.Join(value.dir, value.name)
}
