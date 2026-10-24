package integration_test

import (
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Imported macro context", func() {
	for _, form := range []struct {
		name      string
		statement string
		call      string
	}{
		{"import", `{% import "/macros/greeting" as greetings %}`, `greetings.greet("my friend")`},
		{"from import", `{% from "/macros/greeting" import greet as hello %}`, `hello("my friend")`},
	} {
		Context(form.name, func() {
			var templates map[string]string
			var render func() (string, error)

			BeforeEach(func() {
				templates = map[string]string{
					"/main": form.statement + `<p>{{ ` + form.call + ` }}</p>`,
					"/macros/greeting": `{% set prefix = "Good morning" %}` +
						`{% macro greet(name) %}{{ prefix }}, {{ name }}{% endmacro %}`,
				}
				render = func() (string, error) {
					loader := loaders.MustNewMemoryLoader(templates)
					template, err := exec.NewTemplate("/main", gonja.DefaultConfig, loader, gonja.DefaultEnvironment)
					if err != nil {
						return "", err
					}
					return template.ExecuteToString(nil)
				}
			})

			It("keeps variables defined in the imported template", func() {
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good morning, my friend</p>"))
			})

			It("keeps assignments after the macro definition", func() {
				templates["/macros/greeting"] += `{% set prefix = "Good evening" %}`
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good evening, my friend</p>"))
			})

			It("isolates imported assignments and output from the caller", func() {
				templates["/main"] = `{% set prefix = "Caller" %}` + templates["/main"] + `{{ prefix }}`
				templates["/macros/greeting"] = "discard this output" + templates["/macros/greeting"]
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good morning, my friend</p>Caller"))
			})

			It("resolves sibling macros in the imported template", func() {
				templates["/macros/greeting"] = `{% set prefix = "Good morning" %}` +
					`{% macro greeting() %}{{ prefix }}{% endmacro %}` +
					`{% macro greet(name) %}{{ greeting() }}, {{ name }}{% endmacro %}`
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good morning, my friend</p>"))
			})

			It("resolves nested imports relative to the imported template", func() {
				templates["/macros/greeting"] = `{% from "prefix" import greeting %}` +
					`{% macro greet(name) %}{{ greeting() }}, {{ name }}{% endmacro %}`
				templates["/macros/prefix"] = `{% set prefix = "Good morning" %}` +
					`{% macro greeting() %}{{ prefix }}{% endmacro %}`
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good morning, my friend</p>"))
			})

			It("evaluates macro defaults in the imported context", func() {
				templates["/macros/greeting"] = `{% set prefix = "Good morning" %}` +
					`{% macro greet(name, greeting=prefix) %}{{ greeting }}, {{ name }}{% endmacro %}`
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("<p>Good morning, my friend</p>"))
			})

			It("supports repeated call blocks", func() {
				templates["/macros/greeting"] = `{% macro greet(name) %}{{ caller() }}{% endmacro %}`
				templates["/main"] = form.statement +
					`{% call ` + form.call + ` %}first{% endcall %}` +
					`{% call ` + form.call + ` %}second{% endcall %}`
				result, err := render()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("firstsecond"))
			})

			It("propagates errors when initializing imported variables", func() {
				templates["/macros/greeting"] = `{% set prefix = missing_function() %}` +
					`{% macro greet(name) %}{{ prefix }}, {{ name }}{% endmacro %}`
				_, err := render()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing_function"))
			})
		})
	}
})
