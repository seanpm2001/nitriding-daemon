package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// logRecord
type logRecord struct {
	digestType byte
	digestSize byte
	digest     []byte
	time       time.Time
}

func newLogRecord(blob []byte) *logRecord {
	digest := sha256.Sum256(blob)
	return &logRecord{
		digestType: 0x12, // SHA-2-256.
		digestSize: 0x20, // 32 bytes, in the "variable integer" multiformat.
		digest:     digest[:],
		time:       time.Now().UTC(),
	}
}

func (r *logRecord) String() string {
	return fmt.Sprintf("%s: %x (type=%x)\n", r.time.Format(time.RFC3339), r.digest, r.digestType)
}

// memLog implements a transparencyLog in memory.
type memLog []*logRecord

func (l *memLog) append(r *logRecord) error {
	*l = append(*l, r)
	elog.Printf("Appended %s to transparency log of new size %d.", r, len(*l))
	return nil
}

func (l memLog) String() string {
	var s string
	for _, r := range l {
		elog.Println(r)
		s += r.String()
	}
	return s
}
