package database

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Entry represents a local database entry for the tests
// There are currently 3 buckets: courses, faculties, accounts,
// each bucket contains many of the following entries:
//
// 1) A course entry has ID as the contract address and
// the following map of elements:
// - evaluators: list of evaluators addresses
// - students: list of student addresses
// - credentials: list of credentials hashes
//
// 2) A faculty entry has ID as the contract address and
// the following map of elements:
// - adms: list of adms addresses
// - courses: list of courses contract addresses
// - credentials: list of credentials hashes
//
// 3) A account entry has the ID as the ethereum address
// and the following map of elements:
// - privKey: hex of the private key
type Entry struct {
	ID       string
	Elements map[string]map[string]struct{}
}

func NewEntry(id string, elements ...map[string][]string) *Entry {
	ce := Entry{ID: id}
	ce.Elements = make(map[string]map[string]struct{})
	for _, entries := range elements {
		ce.addElement(entries)
	}
	return &ce
}

func (e *Entry) addElement(elements map[string][]string) {
	for key, list := range elements {
		_, ok := e.Elements[key]
		if !ok {
			e.Elements[key] = make(map[string]struct{})
		}
		e.Elements[key] = AppendUnique(e.Elements[key], list)
	}
}

func (e *Entry) deleteElement(item string) {
	for key, elem := range e.Elements {
		if key == item {
			delete(e.Elements, item)
		} else {
			_, ok := elem[item]
			if ok {
				delete(elem, item)
			}
		}
	}
}

func AppendUnique(a map[string]struct{}, b []string) map[string]struct{} {
	for _, k := range b {
		_, ok := a[k]
		if !ok && k != "" {
			a[k] = struct{}{}
		}
	}
	return a
}

func (e *Entry) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(e)
	if err != nil {
		log.Fatalf("Encode error: %v\n", err)
	}
	return result.Bytes()
}

func DeserializeEntry(buf []byte) *Entry {
	var e Entry
	decoder := gob.NewDecoder(bytes.NewReader(buf))
	err := decoder.Decode(&e)
	if err != nil {
		log.Fatalf("Decode error: %v\n", err)
	}
	return &e
}
