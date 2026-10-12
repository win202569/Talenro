package nodecerttime

import "time"

func certificateWindow(now, start, end time.Time) error {
	if now.IsZero() || now.Year() < 1 || now.Year() > 9999 || start.Year() < 1 || start.Year() > 9999 || end.Year() < 1 || end.Year() > 9999 || !start.Before(end) {
		return ErrInvalidEvidence
	}
	if now.Before(start) {
		return ErrNotYetValid
	}
	if !now.Before(end) {
		return ErrExpired
	}
	return nil
}
func checkWindows(s Summary) (time.Time, error) {
	now, until := s.Time.Policy.Time.Now, s.Time.ValidUntil
	caStart, caEnd := s.Certificate.CA.NotBefore, s.Certificate.CA.NotAfter
	leafStart, leafEnd := s.Certificate.Leaf.Leaf.NotBefore, s.Certificate.Leaf.Leaf.NotAfter
	if now.IsZero() || until.IsZero() || now.Year() < 1 || now.Year() > 9999 || until.Year() < 1 || until.Year() > 9999 || !now.Before(until) ||
		caStart.Year() < 1 || caStart.Year() > 9999 || caEnd.Year() < 1 || caEnd.Year() > 9999 || !caStart.Before(caEnd) ||
		leafStart.Year() < 1 || leafStart.Year() > 9999 || leafEnd.Year() < 1 || leafEnd.Year() > 9999 || !leafStart.Before(leafEnd) {
		return time.Time{}, ErrInvalidEvidence
	}
	if e := certificateWindow(now, caStart, caEnd); e != nil {
		return time.Time{}, e
	}
	if e := certificateWindow(now, leafStart, leafEnd); e != nil {
		return time.Time{}, e
	}
	if caEnd.Before(until) {
		until = caEnd
	}
	if leafEnd.Before(until) {
		until = leafEnd
	}
	return until, nil
}
