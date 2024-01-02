package main

import "html/template"

type TemplateHolder struct {
	templates map[string]*template.Template
}

func (th *TemplateHolder) Get(name string) *template.Template {
	return th.templates[name]
}

func (th *TemplateHolder) Add(name string, template2 template.Template) {
	th.templates[name] = &template2
}
