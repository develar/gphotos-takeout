package main

import (
  "bytes"
  "errors"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/valyala/fastjson"
  "lukechampine.com/blake3"
)

func main() {
  inDir := flag.String("input", "", "The Google Photos take-out directory.")
  outDir := flag.String("output", "", "The output directory.")
  flag.Parse()
  err := layout(*inDir, *outDir)
  if err != nil {
    log.Fatalf("%v", err)
  }
}

func layout(rootDir string, outDir string) error {
  if rootDir == "" {
    return errors.New("input directory is not specified")
  }
  if outDir == "" {
    return errors.New("output directory is not specified")
  }

  log.Printf("start (dir=%s, outDir=%s)\n", rootDir, outDir)

  yearMap := make(map[int]map[time.Month]map[ItemKey]ItemValue)

  var jsonParser fastjson.Parser

  yearOrAlbumNames, err := readDir(rootDir)
  if err != nil {
    return err
  }

  start := time.Now()

  var years []int
  for _, file := range yearOrAlbumNames {
    if strings.HasSuffix(file, ".json") || strings.HasPrefix(file, ".") {
      continue
    }

    yearOrAlbumDir := filepath.Join(rootDir, file)
    //if strings.HasPrefix(file, "All ") ||
    //  strings.HasPrefix(file, "2014-") ||
    //  strings.HasPrefix(file, "2015-") ||
    //  strings.HasPrefix(file, "2016-") ||
    //  strings.HasPrefix(file, "2017-") ||
    //  strings.HasPrefix(file, "2016-") ||
    //  strings.HasPrefix(file, "2018-") ||
    //  strings.HasPrefix(file, "2019-") ||
    //  strings.HasPrefix(file, "Trip to ") ||
    //  strings.HasPrefix(file, "2020-") {
    //  continue
    //}

    log.Print("scan dir: " + file)

    itemNames, err := readDir(yearOrAlbumDir)
    if err != nil {
      return err
    }

    nameMap := make(map[string]bool)
    for _, name := range itemNames {
      nameMap[strings.ToLower(name)] = true
    }

    for _, name := range itemNames {
      if !strings.HasSuffix(name, ".json") {
        continue
      }

      path := filepath.Join(yearOrAlbumDir, name)
      err = processMetaItem(name, path, yearOrAlbumDir, nameMap, &jsonParser, yearMap, &years)
      if err != nil {
        return fmt.Errorf("cannot process %s: %w", path, err)
      }
    }
  }

  log.Printf("scanning complete (%s)", time.Since(start).Round(time.Millisecond).String())

  start = time.Now()

  sort.Ints(years)
  err = os.Mkdir(outDir, 0755)
  if err != nil && !os.IsExist(err) {
    return err
  }

  err = linkItems(years, outDir, yearMap)
  if err != nil {
    return err
  }

  log.Printf("linking complete (%s)", time.Since(start).Round(time.Millisecond).String())
  return nil
}

func readDir(dir string) ([]string, error) {
  f, err := os.Open(dir)
  if err != nil {
    return nil, err
  }

  list, err := f.Readdirnames(-1)
  _ = f.Close()
  if err != nil {
    return nil, err
  }
  // constant order of processing to avoid any random bugs
  sort.Strings(list)
  return list, nil
}

func processMetaItem(
  metaName string,
  path string,
  dir string,
  nameSet map[string]bool,
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
  itemFileName, err := getItemFileName(metaName, nameSet)
  if err != nil {
    log.Printf("WARN: cannot find original file for %s\n", path)
    return nil
  }

  if len(itemFileName) == 0 {
    return errors.New("itemFileName is empty")
  }

  ext := filepath.Ext(itemFileName)
  edited := itemFileName[:len(itemFileName)-len(ext)] + "-edited" + ext
  if !nameSet[edited] {
    edited = ""
  }

  timeMap[itemKey] = ItemValue{
    dir:        dir,
    name:       itemFileName,
    editedName: edited,
  }
  return nil
}

var itemExtensions = []string{"jpg", "gif", "jpeg", "webm", "webp", "mp4", "mov", "mkv"}

func getItemFileName(metaName string, nameMap map[string]bool) (string, error) {
  itemName := strings.ToLower(metaName[:len(metaName)-len(".json")])
  if nameMap[itemName] {
    return itemName, nil
  }

  // for creations with long name jpg is shortened to j
  if strings.HasSuffix(itemName, ".j") {
    itemName += "pg"
    if nameMap[itemName] {
      return itemName, nil
    }
  }

  nameWithoutExtension := strings.ToLower(metaName[:len(metaName)-len("json")])
  for _, ext := range itemExtensions {
    // P1020323.JPG(1). -> P1020323(1).
    itemName := strings.Replace(nameWithoutExtension, "."+ext+"(", "(", 1) + ext
    if nameMap[itemName] {
      return itemName, nil
    }
  }

  return "", fmt.Errorf("cannot determinate image file metaName (metaName=%s)", metaName)
}

func linkItems(years []int, outDir string, yearMap map[int]map[time.Month]map[ItemKey]ItemValue) error {
  for _, year := range years {
    yearDir := filepath.Join(outDir, strconv.Itoa(year))
    err := os.Mkdir(yearDir, 0755)
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

      keys := make([]ItemKey, 0, len(itemMap))
      for key := range itemMap {
        keys = append(keys, key)
      }

      sort.Slice(keys, func(i, j int) bool {
        return keys[i].time.Unix() < keys[j].time.Unix()
      })

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
  outImageFile := filepath.Join(monthDir, item.name)

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
    if !strings.EqualFold(item.name, "color_pop.jpg") &&
      !strings.EqualFold(item.name, "effects.jpg") &&
      !strings.EqualFold(item.name, "movie") &&
      !strings.EqualFold(item.name, "movie(1).mp4") {
      log.Printf("file name is duplicated but content is different, written with suffix %s (previousFile=%s, currentFile=%s, meta=%s)",
        timeSuffix,
        filepath.Join(writtenNames[item.name].dir, item.name),
        filepath.Join(item.dir, item.name),
        filepath.Join(item.dir, key.metaName),
      )
    }

    outImageFile := filepath.Join(monthDir, nameWithSuffix(item.name, timeSuffix))
    err = os.Link(inImageFile, outImageFile)
    if err != nil {
      return err
    }
  }

  // ensure that meta file name is consistent - do not use google take-out original name
  err = os.Link(filepath.Join(item.dir, key.metaName), filepath.Join(monthDir, nameWithSuffix(item.name, timeSuffix)+".json"))
  if err != nil {
    return err
  }

  if len(item.editedName) != 0 {
    err = os.Link(filepath.Join(item.dir, item.editedName), filepath.Join(monthDir, nameWithSuffix(item.name, "_edited" + timeSuffix)))
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

func hashFile(file string) ([]byte, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }

  //goland:noinspection GoUnhandledErrorResult
  defer f.Close()

  h := blake3.New(512, nil)
  if _, err := io.Copy(h, f); err != nil {
    return nil, err
  }
  return h.Sum(nil), nil
}

type ItemKey struct {
  metaName string
  time     time.Time
}

type ItemValue struct {
  dir        string
  name       string
  editedName string
}

func (t *ItemKey) String() string {
  return fmt.Sprintf("ItemKey(metaName=%s, time=%v))", t.metaName, t.time)
}

func (t *ItemValue) String() string {
  return fmt.Sprintf("ItemValue(name=%s, dir=%s, editedName=%s))", t.name, t.dir, t.editedName)
}
