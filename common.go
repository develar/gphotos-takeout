package main

import (
  "facette.io/natsort"
  "fmt"
  "io"
  "lukechampine.com/blake3"
  "os"
  "regexp"
  "sort"
  "time"
)

type ItemKey struct {
  metaName string
  time     time.Time
}

func (t *ItemKey) String() string {
  return fmt.Sprintf("ItemKey(metaName=%s, time=%v))", t.metaName, t.time)
}

type ItemValue struct {
  dir        string
  name       string
  editedName string
}

func (t *ItemValue) String() string {
  return fmt.Sprintf("ItemValue(name=%s, dir=%s, editedName=%s))", t.name, t.dir, t.editedName)
}

var numPrefixRe = regexp.MustCompile(`\(\d+\)`)

func getKeys(itemMap map[ItemKey]ItemValue) []ItemKey {
  keys := make([]ItemKey, 0, len(itemMap))
  for key := range itemMap {
    keys = append(keys, key)
  }

  sort.Slice(keys, func(i, j int) bool {
    a := keys[i]
    b := keys[j]
    if a.time.Equal(b.time) {
      less := natsort.Compare(a.metaName, b.metaName)
      // nat sort puts a(1).jpg before a.jpg
      if less && len(a.metaName) >= (len(b.metaName) + 3) {
        stripped := numPrefixRe.ReplaceAllString(a.metaName, "")
        if stripped == b.metaName {
          return false
        }
      }
      return less
    }
    return a.time.Before(b.time)
  })
  return keys
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
