package models

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrorOfficeNotOpendYet   = errors.New("Office Not Opened Yet")
	ErrorOfficeClosed        = errors.New("Office Closed")
	ErrorInvalidOfficeHours  = errors.New("Invalid Office Hours")
	ErrorEpochReservedBefore = errors.New("Error Reserved Before")

	ErrorBlockAllocation = errors.New("Block Allocation")
)

type Office interface {
	Book(ses Session) (*Session, error)
	IsBookable(Session) bool
	UnBook(BookID string) error
}

func NewOffice(Open, Close string) (Office, error) {
	var err error
	var fixed bool
	o := new(office)

	o.OpenEpoch, err, _ = EpochFromTime(Open)
	if err != nil {
		return nil, err
	}

	o.CloseEpoch, err, fixed = EpochFromTime(Close)
	if err != nil {
		return nil, err
	}
	if fixed {
		o.CloseEpoch--
	}
	o.length = int(o.CloseEpoch-o.OpenEpoch) + 1
	if o.OpenEpoch > o.CloseEpoch {
		return nil, ErrorInvalidOfficeHours
	}

	o.Sessions = make(map[string]*Session)
	o.Epochs = make([]string, o.length, o.length)
	return o, err
}

// var _ Office = office{}

type office struct {
	sync.Mutex

	OpenEpoch  Epoch
	CloseEpoch Epoch
	length     int
	Epochs     []string // ["","","key1","key1","key1","","key2","key2",...]
	Sessions   map[string]*Session
}

func (o *office) IsBookable(ses Session) bool {
	eps := ses.Epochs()
	// isbookable = true;
	for _, v := range eps {
		if o.Epochs[v-o.OpenEpoch] != "" {
			// the session reserved by o.Epochs[v-o.OpenEpoch] SessionKey
			return false
		}
	}
	return true
}

func (o *office) Book(ses Session) (*Session, error) {
	if !o.IsBookable(ses) {
		return nil, ErrorEpochReservedBefore
	}

	s := newSession(ses)

	eps := s.Epochs()
	log.Println(eps)
	o.Lock()

	for _, v := range eps {
		o.Epochs[v-o.OpenEpoch] = s.ID
	}
	o.Sessions[s.ID] = s
	o.Unlock()
	return s, nil
}

func (o *office) UnBook(sesID string) (err error) {
	s, ok := o.Sessions[sesID]
	if !ok {
		// ? should I return a error or not 🧐
		return nil
	}
	eps := s.Epochs()
	o.Lock()
	delete(o.Sessions, sesID)
	for _, v := range eps {
		if o.Epochs[v-o.OpenEpoch] == sesID {
			o.Epochs[v-o.OpenEpoch] = ""
		} else {
			return ErrorBlockAllocation
		}
	}
	o.Unlock()
	return nil
}
