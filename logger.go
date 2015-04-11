package main

import (
	"github.com/Sirupsen/logrus"
)

type wrappedLogger logrus.Logger

func (w *wrappedLogger) Debug(msg string, vars ...interface{}) {
	f := logrus.Fields{}
	for i := 0; i < len(vars)/2; i++ {
		f[vars[i*2].(string)] = vars[i*2+1]
	}
	(*logrus.Logger)(w).WithFields(f).Debug(msg)
}
func (w *wrappedLogger) Error(msg string, vars ...interface{}) {
	f := logrus.Fields{}
	for i := 0; i < len(vars)/2; i++ {
		f[vars[i*2].(string)] = vars[i*2+1]
	}
	(*logrus.Logger)(w).WithFields(f).Error(msg)
}
func (w *wrappedLogger) Info(msg string, vars ...interface{}) {
	f := logrus.Fields{}
	for i := 0; i < len(vars)/2; i++ {
		f[vars[i*2].(string)] = vars[i*2+1]
	}
	(*logrus.Logger)(w).WithFields(f).Info(msg)
}
func (w *wrappedLogger) Warn(msg string, vars ...interface{}) {
	f := logrus.Fields{}
	for i := 0; i < len(vars)/2; i++ {
		f[vars[i*2].(string)] = vars[i*2+1]
	}
	(*logrus.Logger)(w).WithFields(f).Warn(msg)
}
