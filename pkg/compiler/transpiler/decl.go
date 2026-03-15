package transpiler

import (
	"fmt"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/interpreter"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

func (v *MyGoTranspiler) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{} {
	enumName := types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier())
	defParams, usageParams := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym := v.CurrentScope.Resolve(ctx.ID().GetText()); sym != nil && len(sym.GenericParams) > 0 {
		defParams, usageParams = types.ParseGenericParamMeta(sym.GenericParams)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s%s interface {\n\tis%s()\n}\n", enumName, defParams, enumName))

	// 2. Variants
	for _, variantCtx := range ctx.AllEnumVariant() {
		variantName := variantCtx.ID().GetText()
		fullVariantName := fmt.Sprintf("%s_%s", enumName, variantName)

		// Struct definition
		sb.WriteString(fmt.Sprintf("type %s%s struct {\n", fullVariantName, defParams))

		if variantCtx.TypeList() != nil {
			typesCtx := variantCtx.TypeList().(*ast.TypeListContext).AllTypeType()
			for i, tCtx := range typesCtx {
				typeName := v.resolveType(tCtx)
				sb.WriteString(fmt.Sprintf("\tItem%d %s\n", i+1, typeName))
			}
		}
		sb.WriteString("}\n")

		// Interface implementation
		receiverType := fullVariantName + usageParams
		sb.WriteString(fmt.Sprintf("func (%s) is%s() {}\n", receiverType, enumName))
	}

	// 3. Generate _unwrap helper if variants follow Ok/Err pattern
	var okVariant, errVariant string
	var okType string
	for _, variantCtx := range ctx.AllEnumVariant() {
		vName := variantCtx.ID().GetText()
		if strings.ToLower(vName) == "ok" {
			okVariant = vName
			if variantCtx.TypeList() != nil {
				typesCtx := variantCtx.TypeList().(*ast.TypeListContext).AllTypeType()
				if len(typesCtx) > 0 {
					okType = v.resolveType(typesCtx[0])
				}
			}
		} else if strings.ToLower(vName) == "err" {
			errVariant = vName
		}
	}

	if okVariant != "" && errVariant != "" && okType != "" {
		sb.WriteString(fmt.Sprintf("\nfunc %s_unwrap%s(r %s%s) %s {\n", enumName, defParams, enumName, usageParams, okType))
		sb.WriteString("\tswitch v := r.(type) {\n")
		sb.WriteString(fmt.Sprintf("\tcase %s_%s%s:\n", enumName, okVariant, usageParams))
		sb.WriteString("\t\treturn v.Item1\n")
		sb.WriteString(fmt.Sprintf("\tcase %s_%s%s:\n", enumName, errVariant, usageParams))
		sb.WriteString("\t\tpanic(v.Item1)\n")
		sb.WriteString("\tdefault:\n")
		sb.WriteString("\t\tpanic(\"invalid variant\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString("}\n")
	}

	return sb.String()
}

func (v *MyGoTranspiler) VisitAnnotationDecl(ctx *ast.AnnotationDeclContext) interface{} {
	// Macros are compile-time constructs and should not appear in the output Go code
	return ""
}

func (v *MyGoTranspiler) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	structName := ctx.ID().GetText()
	goStructName := types.FormatVisibility(structName, ctx.Modifier())

	// Resolve or Define symbol to store annotations
	sym := v.CurrentScope.Resolve(structName)
	if sym == nil {
		sym = v.CurrentScope.Define(structName, goStructName, symbols.KindStruct, "")
	}
	// Clear existing annotations to avoid duplication
	sym.Annotations = nil

	// RFC-007: Handle Annotations
	var activeAnnotations []string
	for _, annCtx := range ctx.AllAnnotationUsage() {
		annName := annCtx.ID().GetText()
		var args []string
		if exprList := annCtx.ExprList(); exprList != nil {
			for _, expr := range exprList.(*ast.ExprListContext).AllExpr() {
				args = append(args, expr.GetText())
			}
		}

		// Add to symbol
		ann := &symbols.Annotation{Name: annName, Args: args}
		sym.Annotations = append(sym.Annotations, ann)

		if annName == "Derive" {
			for _, arg := range args {
				// Simple check for "Json" identifier
				if arg == "Json" {
					activeAnnotations = append(activeAnnotations, "Json")
				}
			}
		} else if annName == "Builder" {
			activeAnnotations = append(activeAnnotations, "Builder")
		}
	}

	v.CurrentStructAnnotations[structName] = activeAnnotations
	v.CurrentProcessingStruct = structName
	defer func() {
		v.CurrentProcessingStruct = ""
	}()

	defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym != nil && len(sym.GenericParams) > 0 {
		defParams, _ = types.ParseGenericParamMeta(sym.GenericParams)
	}
	var fields []string
	for _, fCtx := range ctx.AllStructField() {
		if fc, ok := fCtx.(*ast.StructFieldContext); ok {
			res := v.VisitStructField(fc)
			if s, ok := res.(string); ok {
				fields = append(fields, "\t"+s)
			}
		}
	}

	// RFC-008: Execute Macros
	var macroOutputs []string
	for _, annCtx := range ctx.AllAnnotationUsage() {
		annName := annCtx.ID().GetText()

		// Handle @Derive(MacroName)
		if annName == "Derive" && annCtx.ExprList() != nil {
			for _, expr := range annCtx.ExprList().(*ast.ExprListContext).AllExpr() {
				macroName := expr.GetText()

				targetMacroName := "Derive" + macroName
				macroSym := v.CurrentScope.Resolve(targetMacroName)

				// DEBUG
				kind := "nil"
				if macroSym != nil {
					kind = string(macroSym.Kind)
				}
				fmt.Printf("DEBUG: Derive check: %s -> %s found=%v kind=%s\n", macroName, targetMacroName, macroSym != nil, kind)

				if macroSym != nil && macroSym.Kind == symbols.KindAnnotation { // Annotation/Macro
					// Found macro!
					// Execute it
					if macroNode, ok := macroSym.ASTNode.(*ast.AnnotationDeclContext); ok {
						interp := interpreter.NewInterpreter(v.CurrentScope)
						// Create target meta
						targetMeta := interp.CreateSymbolMeta(sym)
						interp.Env.Set("target", targetMeta)

						// Execute Block
						res := interp.VisitBlock(macroNode.Block().(*ast.BlockContext))
						if ret, ok := res.(interpreter.ReturnValue); ok {
							if strVal, ok := ret.Val.(interpreter.StringValue); ok {
								goCode := v.transpileMacroResult(strVal.Val)
								macroOutputs = append(macroOutputs, goCode)
							}
						}
					}
				}
			}
		}
	}

	return fmt.Sprintf("type %s%s struct {\n%s\n}\n\n%s", goStructName, defParams, strings.Join(fields, "\n"), strings.Join(macroOutputs, "\n\n"))
}

func (v *MyGoTranspiler) VisitStructField(ctx *ast.StructFieldContext) interface{} {
	fieldName := ctx.ID().GetText()
	fieldType := types.ResolveTypeWithScope(ctx.TypeType().GetText(), v.CurrentScope)
	typeName := v.toGoType(fieldType)

	tag := ""
	rawTag := ""
	if ctx.STRING() != nil {
		rawTag = ctx.STRING().GetText()
		tag = " " + rawTag
	}
	if specializedType, ok := v.specializedTaggedOptionalFieldGoType(fieldType, rawTag); ok {
		typeName = specializedType
	}

	return fmt.Sprintf("%s %s%s", fieldName, typeName, tag)
}

func (v *MyGoTranspiler) VisitPureTraitDecl(ctx *ast.PureTraitDeclContext) interface{} {
	traitName := ctx.ID().GetText()
	goTraitName := types.FormatVisibility(traitName, ctx.Modifier())

	defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym := v.CurrentScope.Resolve(traitName); sym != nil && len(sym.GenericParams) > 0 {
		defParams, _ = types.ParseGenericParamMeta(sym.GenericParams)
	}

	var methods []string
	for _, fn := range ctx.AllTraitFnDecl() {
		fnName := fn.ID().GetText()

		// Parse params
		var params []string
		if fn.ParamList() != nil {
			for _, p := range fn.ParamList().(*ast.ParamListContext).AllParam() {
				// name := p.ID().GetText()
				typ := v.resolveType(p.TypeType())
				params = append(params, typ)
			}
		}

		// Parse return type
		retType := ""
		if fn.TypeType() != nil {
			retType = v.resolveType(fn.TypeType())
		}

		sig := fmt.Sprintf("%s(%s)", fnName, strings.Join(params, ", "))
		if retType != "" {
			sig += " " + retType
		}
		methods = append(methods, "\t"+sig)
	}

	return fmt.Sprintf("type %s%s interface {\n%s\n}\n", goTraitName, defParams, strings.Join(methods, "\n"))
}

func (v *MyGoTranspiler) VisitBindTraitDecl(ctx *ast.BindTraitDeclContext) interface{} {
	// Collect directives
	for _, item := range ctx.AllTraitBodyItem() {
		if item.CompositionDirective() != nil {
			// Process ban/flip ban
			_ = item.CompositionDirective().Accept(v)
		}
	}
	return "" // TODO: Implement bind trait logic correctly
}

func (v *MyGoTranspiler) VisitBanDirective(ctx *ast.BanDirectiveContext) interface{} {
	var methods []string
	for _, id := range ctx.AllID() {
		methods = append(methods, id.GetText())
	}
	return methods
}

type FlipBanItem struct {
	Method string
	Trait  string
}

func (v *MyGoTranspiler) VisitFlipBanDirective(ctx *ast.FlipBanDirectiveContext) interface{} {
	var items []FlipBanItem
	for _, itemCtx := range ctx.AllFlipBanItem() {
		if item, ok := itemCtx.Accept(v).(FlipBanItem); ok {
			items = append(items, item)
		}
	}
	return items
}

func (v *MyGoTranspiler) VisitFlipBanItem(ctx *ast.FlipBanItemContext) interface{} {
	ids := ctx.AllID()
	if len(ids) == 2 {
		return FlipBanItem{
			Method: ids[0].GetText(),
			Trait:  ids[1].GetText(),
		}
	}
	return nil
}

func (v *MyGoTranspiler) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	fnName := ctx.ID().GetText()

	oldFnName := v.CurrentOriginalFnName
	v.CurrentOriginalFnName = fnName
	defer func() { v.CurrentOriginalFnName = oldFnName }()

	var params []string
	var receiver string
	if ctx.ParamList() != nil {
		for i, p := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			pName := p.ID().GetText()
			pType := v.resolveType(p.TypeType())

			if i == 0 && (pName == "self" || pName == "this") {
				// It's a method!
				receiver = fmt.Sprintf("(%s %s)", pName, pType)
			} else {
				params = append(params, fmt.Sprintf("%s %s", pName, pType))
			}
		}
	}

	retType := ""
	if ctx.TypeType() != nil {
		retType = v.resolveType(ctx.TypeType())
	}

	v.pushScope("fn_" + fnName)
	if ctx.ParamList() != nil {
		for _, p := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			pName := p.ID().GetText()
			v.CurrentScope.Define(pName, pName, symbols.KindVar, "unknown")
		}
	}

	body := ""
	if ctx.Block() != nil {
		if b, ok := ctx.Block().(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				body = s
			}
		}
	}

	v.popScope()

	sig := ""
	if receiver != "" {
		sig = fmt.Sprintf("func %s %s(%s)", receiver, fnName, strings.Join(params, ", "))
	} else {
		sig = fmt.Sprintf("func %s(%s)", fnName, strings.Join(params, ", "))
	}

	if retType != "" {
		sig += " " + retType
	}

	return sig + " " + body
}
