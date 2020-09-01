package main

import (
  "facette.io/natsort"
  "fmt"
  "log"
  "os"
  "path/filepath"
  "sort"
  "strings"
  "time"

  "github.com/spf13/cobra"
  "github.com/valyala/fastjson"
)

func main() {
  var input string
  var outDir string
  var dirFilter string

  var rootCmd = &cobra.Command{
    Use: "gphotos-takeout",
    RunE: func(cmd *cobra.Command, args []string) error {
      return layout(input, outDir, dirFilter)
    },
  }

  rootCmd.Flags().StringVarP(&input, "input", "i", "", "path to Takeout\\Google Photos directory")
  err := rootCmd.MarkFlagRequired("input")
  if err != nil {
    panic(err)
  }

  rootCmd.Flags().StringVarP(&outDir, "output", "o", "", "path to output directory")
  err = rootCmd.MarkFlagRequired("output")
  if err != nil {
    panic(err)
  }

  rootCmd.Flags().StringVarP(&dirFilter, "filter", "f", "", "directory names pattern (e.g. '2020-08-2*')")

  err = rootCmd.Execute()
  if err != nil {
    log.Fatal(err)
  }
}

func layout(rootDir string, outDir string, dirFilter string) error {
  log.Printf("start (dir=%s, dirFilter=%s, outDir=%s)\n", rootDir, dirFilter, outDir)

  yearMap := make(map[int]map[time.Month]map[ItemKey]ItemValue)

  var jsonParser fastjson.Parser

  yearOrAlbumNames, err := readDir(rootDir)
  // constant order of processing to avoid any random bugs
  // nat sort not suitable - albums should be always after year dirs
  sort.Strings(yearOrAlbumNames)
  if err != nil {
    return err
  }

  start := time.Now()

  var years []int
  for _, file := range yearOrAlbumNames {
    if strings.HasSuffix(file, ".json") || strings.HasPrefix(file, ".") {
      continue
    }

    if dirFilter != "" {
      match, err := filepath.Match(dirFilter, file)
      if err != nil {
        return err
      }
      if !match {
        continue
      }
    }

    yearOrAlbumDir := filepath.Join(rootDir, file)
    log.Print("scan dir: " + file)

    itemNames, err := readDir(yearOrAlbumDir)
    // for items use nat sort
    natsort.Sort(itemNames)
    if err != nil {
      return err
    }

    nameMap := make(map[string]string)
    for _, name := range itemNames {
      nameMap[strings.ToLower(name)] = name
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

  sort.Ints(years)

  start = time.Now()
  err = deduplicate(years, yearMap, rootDir)
  if err != nil {
    return err
  }

  log.Printf("deduplicating complete (%s)", time.Since(start).Round(time.Millisecond).String())

  start = time.Now()
  err = linkItems(years, yearMap, outDir)
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
  return list, nil
}
