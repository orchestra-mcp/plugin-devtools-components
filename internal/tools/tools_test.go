package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// callTool builds a ToolRequest from args and invokes the handler.
func callTool(t *testing.T, handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error), args map[string]any) *pluginv1.ToolResponse {
	t.Helper()
	s, err := structpb.NewStruct(args)
	if err != nil {
		t.Fatalf("callTool: structpb.NewStruct: %v", err)
	}
	req := &pluginv1.ToolRequest{Arguments: s}
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("callTool: handler returned error: %v", err)
	}
	return resp
}

func isError(resp *pluginv1.ToolResponse) bool {
	return resp != nil && !resp.Success
}

func errorCode(resp *pluginv1.ToolResponse) string {
	return resp.GetErrorCode()
}

// responseText extracts the "text" field from the Result struct.
func responseText(resp *pluginv1.ToolResponse) string {
	if resp == nil || resp.Result == nil {
		return ""
	}
	v, ok := resp.Result.Fields["text"]
	if !ok || v == nil {
		return ""
	}
	sv, ok := v.Kind.(*structpb.Value_StringValue)
	if !ok {
		return ""
	}
	return sv.StringValue
}

// makeTempComponents creates a temp directory with Button.tsx, Card.vue, and Toggle.svelte.
func makeTempComponents(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"Button.tsx": `import React from 'react';
interface ButtonProps { label: string; }
export const Button: React.FC<ButtonProps> = ({ label }) => <button>{label}</button>;
export default Button;
`,
		"Card.vue": `<template><div class="card"><slot /></div></template>
<script>
export default { name: 'Card', props: { title: String } };
</script>
<style scoped></style>
`,
		"Toggle.svelte": `<script>
export let checked = false;
</script>
<input type="checkbox" bind:checked />
`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("makeTempComponents: write %s: %v", name, err)
		}
	}
	return dir
}

// ---------------------------------------------------------------------------
// ComponentList
// ---------------------------------------------------------------------------

func TestComponentList_MissingDirectory(t *testing.T) {
	handler := ComponentList()
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentList_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentList()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	if !strings.Contains(text, "No component") {
		t.Fatalf("expected 'No component' in response, got: %q", text)
	}
}

func TestComponentList_WithComponents(t *testing.T) {
	dir := makeTempComponents(t)
	handler := ComponentList()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	if !strings.Contains(text, "Button") {
		t.Fatalf("expected 'Button' in response, got: %q", text)
	}
	if !strings.Contains(text, "Card") {
		t.Fatalf("expected 'Card' in response, got: %q", text)
	}
	if !strings.Contains(text, "Toggle") {
		t.Fatalf("expected 'Toggle' in response, got: %q", text)
	}
}

func TestComponentList_WithFilter(t *testing.T) {
	dir := makeTempComponents(t)
	handler := ComponentList()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"filter":    "Button",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	if !strings.Contains(text, "Button") {
		t.Fatalf("expected 'Button' in response, got: %q", text)
	}
	if strings.Contains(text, "Card") {
		t.Fatalf("expected 'Card' to be filtered out, got: %q", text)
	}
	if strings.Contains(text, "Toggle") {
		t.Fatalf("expected 'Toggle' to be filtered out, got: %q", text)
	}
}

// ---------------------------------------------------------------------------
// ComponentInspect
// ---------------------------------------------------------------------------

func TestComponentInspect_MissingArgs(t *testing.T) {
	handler := ComponentInspect()
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentInspect_NotFound(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentInspect()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "NonExistent",
	})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "not_found" {
		t.Fatalf("expected not_found, got %q", errorCode(resp))
	}
}

func TestComponentInspect_Found(t *testing.T) {
	dir := t.TempDir()
	content := `import React from 'react';
import { useState } from 'react';

interface ButtonProps { label: string; }

export const Button: React.FC<ButtonProps> = ({ label }) => {
  return <button>{label}</button>;
};
export default Button;
`
	if err := os.WriteFile(filepath.Join(dir, "Button.tsx"), []byte(content), 0644); err != nil {
		t.Fatalf("write Button.tsx: %v", err)
	}

	handler := ComponentInspect()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "Button",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s (%s)", resp.ErrorMessage, resp.ErrorCode)
	}
	text := responseText(resp)
	if !strings.Contains(text, "Button") {
		t.Fatalf("expected 'Button' in response, got: %q", text)
	}
}

// ---------------------------------------------------------------------------
// ComponentCreate
// ---------------------------------------------------------------------------

func TestComponentCreate_MissingArgs(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentCreate()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentCreate_ReactDefault(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentCreate()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "Button",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	expectedPath := filepath.Join(dir, "Button.tsx")
	if !strings.Contains(text, expectedPath) {
		t.Fatalf("expected path %q in response, got: %q", expectedPath, text)
	}
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedPath)
	}
}

func TestComponentCreate_Vue(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentCreate()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "MyCard",
		"framework": "vue",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	expectedPath := filepath.Join(dir, "MyCard.vue")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedPath)
	}
	text := responseText(resp)
	if !strings.Contains(text, expectedPath) {
		t.Fatalf("expected path %q in response, got: %q", expectedPath, text)
	}
}

func TestComponentCreate_Svelte(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentCreate()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "Toggle",
		"framework": "svelte",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	expectedPath := filepath.Join(dir, "Toggle.svelte")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedPath)
	}
	text := responseText(resp)
	if !strings.Contains(text, expectedPath) {
		t.Fatalf("expected path %q in response, got: %q", expectedPath, text)
	}
}

// ---------------------------------------------------------------------------
// ComponentPreview
// ---------------------------------------------------------------------------

func TestComponentPreview_MissingPath(t *testing.T) {
	handler := ComponentPreview()
	// ComponentPreview requires "directory" and "name"
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentPreview_NotFound(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentPreview()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "NonExistent",
	})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "not_found" {
		t.Fatalf("expected not_found, got %q", errorCode(resp))
	}
}

func TestComponentPreview_Found(t *testing.T) {
	dir := t.TempDir()
	fileContent := `import React from 'react';
const MyWidget = () => <div>Hello</div>;
export default MyWidget;
`
	if err := os.WriteFile(filepath.Join(dir, "MyWidget.tsx"), []byte(fileContent), 0644); err != nil {
		t.Fatalf("write MyWidget.tsx: %v", err)
	}

	handler := ComponentPreview()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"name":      "MyWidget",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	if !strings.Contains(text, "MyWidget") {
		t.Fatalf("expected component name in response, got: %q", text)
	}
	if !strings.Contains(text, "Hello") {
		t.Fatalf("expected file content in response, got: %q", text)
	}
}

// ---------------------------------------------------------------------------
// ComponentLibrary
// ---------------------------------------------------------------------------

func TestComponentLibrary_MissingDirectory(t *testing.T) {
	handler := ComponentLibrary()
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentLibrary_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentLibrary()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	if !strings.Contains(text, "No components") {
		t.Fatalf("expected 'No components' in response, got: %q", text)
	}
}

func TestComponentLibrary_WithComponents(t *testing.T) {
	dir := makeTempComponents(t)
	handler := ComponentLibrary()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s", resp.ErrorMessage)
	}
	text := responseText(resp)
	// Should contain markdown table header
	if !strings.Contains(text, "| Component |") {
		t.Fatalf("expected markdown table in response, got: %q", text)
	}
	if !strings.Contains(text, "Button") {
		t.Fatalf("expected 'Button' in response, got: %q", text)
	}
	if !strings.Contains(text, "Card") {
		t.Fatalf("expected 'Card' in response, got: %q", text)
	}
	if !strings.Contains(text, "Toggle") {
		t.Fatalf("expected 'Toggle' in response, got: %q", text)
	}
}

// ---------------------------------------------------------------------------
// ComponentSyncFigma
// ---------------------------------------------------------------------------

func TestComponentSyncFigma_MissingDirectory(t *testing.T) {
	handler := ComponentSyncFigma()
	resp := callTool(t, handler, map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error response")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestComponentSyncFigma_NoFileKey(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentSyncFigma()
	resp := callTool(t, handler, map[string]any{"directory": dir})
	if isError(resp) {
		t.Fatalf("expected success (hint), got error: %s (%s)", resp.ErrorMessage, resp.ErrorCode)
	}
	text := responseText(resp)
	if text == "" {
		t.Fatal("expected hint message in response, got empty string")
	}
}

func TestComponentSyncFigma_WithFileKey(t *testing.T) {
	dir := t.TempDir()
	handler := ComponentSyncFigma()
	resp := callTool(t, handler, map[string]any{
		"directory": dir,
		"file_key":  "abc123figmakey",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s (%s)", resp.ErrorMessage, resp.ErrorCode)
	}
}
