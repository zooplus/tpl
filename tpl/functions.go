package tpl

import "text/template"

func (tp *TemplateProcessor) funcMap() template.FuncMap {
	return template.FuncMap{
		"include":     tp.include,
		"mustInclude": tp.mustInclude,
	}
}

func (tp *TemplateProcessor) include(fileName string) (string, error) {
	return tp.renderInclude(fileName, true)
}

func (tp *TemplateProcessor) mustInclude(fileName string) (string, error) {
	return tp.renderInclude(fileName, false)
}
