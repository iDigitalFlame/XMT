package net

type Server interface {
	Sessions() []*Session
	
}
type Session struct{}