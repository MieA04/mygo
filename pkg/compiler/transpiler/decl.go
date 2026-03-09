package transpiler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
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

func (v *MyGoTranspiler) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	goStructName := types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier())
	defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym := v.CurrentScope.Resolve(ctx.ID().GetText()); sym != nil && len(sym.GenericParams) > 0 {
		defParams, _ = types.ParseGenericParamMeta(sym.GenericParams)
	}
	var fields []string
	for _, fCtx := range ctx.AllStructField() {
		fields = append(fields, "\t"+fCtx.Accept(v).(string))
	}
	return fmt.Sprintf("type %s%s struct {\n%s\n}", goStructName, defParams, strings.Join(fields, "\n"))
}

func (v *MyGoTranspiler) VisitPureTraitDecl(ctx *ast.PureTraitDeclContext) interface{} {
	traitName := types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier())
	defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym := v.CurrentScope.Resolve(ctx.ID().GetText()); sym != nil && len(sym.GenericParams) > 0 {
		defParams, _ = types.ParseGenericParamMeta(sym.GenericParams)
	}

	var methods []string
	for _, fnCtx := range ctx.AllTraitFnDecl() {
		methods = append(methods, fnCtx.Accept(v).(string))
	}

	return fmt.Sprintf("type %s%s interface {\n%s\n}", traitName, defParams, strings.Join(methods, "\n"))
}

func (v *MyGoTranspiler) VisitTraitFnDecl(ctx *ast.TraitFnDeclContext) interface{} {
	fnName := ctx.ID().GetText()
	goFnName := fnName // Use name directly to support Go-style visibility (Upper=Public, Lower=Private)

	var params []string
	if ctx.ParamList() != nil {
		for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			pName := p.ID().GetText()
			pType := v.resolveType(p.TypeType())
			params = append(params, fmt.Sprintf("%s %s", pName, pType))
		}
	}

	returnTypeStr := ""
	if ctx.TypeType() != nil {
		returnTypeStr = " " + v.resolveType(ctx.TypeType())
	}

	// If we are in a bind block and have a block, generate a full Go method
	if v.currentImplType != "" && ctx.Block() != nil {
		v.pushScope("trait_fn_" + fnName)
		defer v.popScope()

		// Re-define params in scope for the block
		if ctx.ParamList() != nil {
			for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
				p := pCtx.(*ast.ParamContext)
				pName := p.ID().GetText()
				pType := v.resolveType(p.TypeType())
				v.CurrentScope.Define(pName, pName, symbols.KindVar, pType)
			}
		}

		// Handle bound instance
		receiverStr := ""
		if v.currentBindVar != "" {
			v.CurrentScope.Define(v.currentBindVar, v.currentBindVar, symbols.KindVar, v.currentImplType)
			if v.currentImplSymbol != nil && v.currentImplSymbol.Kind == symbols.KindStruct {
				receiverStr = fmt.Sprintf("(%s *%s) ", v.currentBindVar, v.currentImplType)
			} else {
				instanceParam := fmt.Sprintf("%s %s", v.currentBindVar, v.currentImplType)
				params = append([]string{instanceParam}, params...)
			}
		}

		blockStr := ctx.Block().Accept(v).(string)
		defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
		// Removed incorrect appending of bind generics to method generics
		// if v.currentBindTypeDef != "" { ... }

		if receiverStr != "" {
			return fmt.Sprintf("func %s%s%s(%s)%s %s", receiverStr, goFnName, defParams, strings.Join(params, ", "), returnTypeStr, blockStr)
		}
		return fmt.Sprintf("func %s%s(%s)%s %s", goFnName, defParams, strings.Join(params, ", "), returnTypeStr, blockStr)
	}

	// Otherwise, it's an interface method signature
	return fmt.Sprintf("\t%s(%s)%s", goFnName, strings.Join(params, ", "), returnTypeStr)
}

func (v *MyGoTranspiler) VisitStructField(ctx *ast.StructFieldContext) interface{} {
	return fmt.Sprintf("%s %s", ctx.ID().GetText(), v.resolveType(ctx.TypeType()))
}

func (v *MyGoTranspiler) VisitBindTraitDecl(ctx *ast.BindTraitDeclContext) interface{} {
	var sb strings.Builder
	bindTypeParams := types.ExtractGenericParamMeta(ctx.TypeParams(), v.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), v.CurrentScope)
	mergedMeta, _ := types.MergeGenericConstraints(bindTypeParams, whereMeta)
	v.currentBindTypeDef, _ = types.ParseGenericParamMeta(mergedMeta)
	for _, targetCtx := range ctx.AllBindTarget() {
		tCtx := targetCtx.TypeType()
		if tCtx == nil {
			continue
		}
		targetType := v.resolveType(tCtx)
		v.currentImplType = targetType
		baseName := types.SplitBaseType(tCtx.GetText())
		v.currentImplSymbol = v.CurrentScope.Resolve(baseName)
		if v.currentImplSymbol == nil {
			v.currentImplSymbol = v.CurrentScope.ResolveByGoName(baseName)
		}
		v.currentBindVar = "this"
		if targetCtx.ID() != nil {
			v.currentBindVar = targetCtx.ID().GetText()
		}

		methods, unresolvedConflicts := v.collectBindMethods(ctx)
		if len(unresolvedConflicts) > 0 {
			fmt.Printf("Compile Error: Type '%s' has unresolved trait method conflicts: %s\n", targetType, strings.Join(unresolvedConflicts, ", "))
		}
		for _, fnDecl := range methods {
			sb.WriteString(fnDecl.Accept(v).(string) + "\n")
		}
	}

	v.currentImplType = ""
	v.currentBindVar = ""
	v.currentImplSymbol = nil
	v.currentBindTypeDef = ""
	return sb.String()
}

func (v *MyGoTranspiler) collectBindMethods(ctx *ast.BindTraitDeclContext) ([]*ast.TraitFnDeclContext, []string) {
	explicit := make(map[string]*ast.TraitFnDeclContext)
	var explicitOrder []string
	for _, item := range ctx.AllTraitBodyItem() {
		if fnCtx := item.TraitFnDecl(); fnCtx != nil {
			fnDecl := fnCtx.(*ast.TraitFnDeclContext)
			if fnDecl.Block() == nil {
				continue
			}
			name := fnDecl.ID().GetText()
			if _, ok := explicit[name]; !ok {
				explicitOrder = append(explicitOrder, name)
			}
			explicit[name] = fnDecl
		}
	}

	mixed := make(map[string]*ast.TraitFnDeclContext)
	conflicts := make(map[string]struct{})
	for _, traitSym := range v.resolveCombsTraits(ctx) {
		var names []string
		for name := range traitSym.ConcreteTraitMethods {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if _, exists := mixed[name]; exists {
				conflicts[name] = struct{}{}
				continue
			}
			if fnDecl, ok := traitSym.ConcreteTraitMethods[name].(*ast.TraitFnDeclContext); ok && fnDecl.Block() != nil {
				mixed[name] = fnDecl
			}
		}
	}

	for _, item := range ctx.AllTraitBodyItem() {
		banCtx := item.BanDirective()
		if banCtx == nil {
			continue
		}
		switch directive := banCtx.(type) {
		case *ast.SpecificBanContext:
			text := directive.GetText()
			if strings.HasPrefix(text, "flipban[") {
				keep := make(map[string]struct{})
				for _, id := range directive.AllID() {
					keep[id.GetText()] = struct{}{}
				}
				for name := range mixed {
					if _, ok := keep[name]; ok {
						continue
					}
					delete(mixed, name)
					delete(conflicts, name)
				}
				for name := range conflicts {
					if _, ok := keep[name]; !ok {
						delete(conflicts, name)
					}
				}
				continue
			}
			for _, id := range directive.AllID() {
				name := id.GetText()
				delete(mixed, name)
				delete(conflicts, name)
			}
		case *ast.RepeatBanContext:
			for name := range conflicts {
				delete(mixed, name)
			}
			conflicts = make(map[string]struct{})
		}
	}

	for name := range conflicts {
		if _, ok := explicit[name]; ok {
			delete(conflicts, name)
			continue
		}
		delete(mixed, name)
	}

	var results []*ast.TraitFnDeclContext
	for _, name := range explicitOrder {
		results = append(results, explicit[name])
	}
	var mixedNames []string
	for name := range mixed {
		mixedNames = append(mixedNames, name)
	}
	sort.Strings(mixedNames)
	for _, name := range mixedNames {
		if _, ok := explicit[name]; ok {
			continue
		}
		results = append(results, mixed[name])
	}
	var unresolved []string
	for name := range conflicts {
		unresolved = append(unresolved, name)
	}
	sort.Strings(unresolved)
	return results, unresolved
}

func (v *MyGoTranspiler) resolveCombsTraits(ctx *ast.BindTraitDeclContext) []*symbols.Symbol {
	var traits []*symbols.Symbol
	seen := make(map[string]struct{})
	for _, id := range ctx.AllID() {
		traitName := id.GetText()
		traitSym := v.CurrentScope.Resolve(traitName)
		if traitSym == nil || traitSym.Kind != symbols.KindTrait {
			continue
		}
		if _, ok := seen[traitSym.MyGoName]; ok {
			continue
		}
		seen[traitSym.MyGoName] = struct{}{}
		traits = append(traits, traitSym)
	}
	return traits
}

func (v *MyGoTranspiler) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	fnName := ctx.ID().GetText()
	goFnName := types.FormatVisibility(fnName, ctx.Modifier())
	defParams, _ := types.ParseTypeParams(ctx.TypeParams(), v.CurrentScope)
	if sym := v.CurrentScope.Resolve(fnName); sym != nil && len(sym.GenericParams) > 0 {
		defParams, _ = types.ParseGenericParamMeta(sym.GenericParams)
	}

	v.pushScope("fn_" + fnName)

	isMethod := v.currentImplType != ""
	receiverStr := ""

	var isStructMethod bool

	_, implTypeParamsStr := types.InferImplTypeParamDefs(v.currentImplType, v.Scope)
	if v.currentBindTypeDef != "" {
		implTypeParamsStr = v.currentBindTypeDef
	}

	if isMethod && v.currentImplSymbol != nil {
		if v.currentImplSymbol.Kind == symbols.KindEnum {
			// Enum methods are standalone functions
			goFnName = fmt.Sprintf("%s_%s", v.currentImplSymbol.GoName, fnName)
		} else if v.currentImplSymbol.Kind == symbols.KindStruct {
			// Struct methods use receiver
			isStructMethod = true
		}
	}

	// Merge generic params for non-struct methods (e.g. Enum standalone functions)
	if implTypeParamsStr != "" && !isStructMethod {
		if defParams == "" {
			defParams = implTypeParamsStr
		} else {
			p1 := implTypeParamsStr[1 : len(implTypeParamsStr)-1]
			p2 := defParams[1 : len(defParams)-1]
			defParams = "[" + p1 + ", " + p2 + "]"
		}
	}

	var params []string
	if ctx.ParamList() != nil {
		for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			pName := p.ID().GetText()
			pType := v.resolveType(p.TypeType())
			params = append(params, fmt.Sprintf("%s %s", pName, pType))
			v.CurrentScope.Define(pName, pName, symbols.KindVar, pType)
		}
	}

	// Inject bound instance
	if v.currentBindVar != "" && v.currentImplType != "" {
		// Define instance in scope
		v.CurrentScope.Define(v.currentBindVar, v.currentBindVar, symbols.KindVar, v.currentImplType)

		if isStructMethod {
			// Receiver for Struct (Always use pointer receiver for mutable semantics)
			receiverStr = fmt.Sprintf("(%s *%s) ", v.currentBindVar, v.currentImplType)
		} else {
			// First parameter for Enum (or others)
			instanceParam := fmt.Sprintf("%s %s", v.currentBindVar, v.currentImplType)
			params = append([]string{instanceParam}, params...)
		}
	}

	returnTypeStr := ""
	if ctx.TypeType() != nil {
		returnTypeStr = " " + v.resolveType(ctx.TypeType())
	}
	blockStr := ctx.Block().Accept(v).(string)
	v.popScope()

	// Helper to handle receiver placement
	if receiverStr != "" {
		return fmt.Sprintf("func %s%s%s(%s)%s %s", receiverStr, goFnName, defParams, strings.Join(params, ", "), returnTypeStr, blockStr)
	}
	return fmt.Sprintf("func %s%s(%s)%s %s", goFnName, defParams, strings.Join(params, ", "), returnTypeStr, blockStr)
}

func (v *MyGoTranspiler) VisitSingleLetDecl(ctx *ast.SingleLetDeclContext) interface{} {
	varName := types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier())

	declaredType := ""
	if ctx.TypeType() != nil {
		declaredType = v.resolveType(ctx.TypeType())
	}

	oldExpected := v.expectedType
	defer func() { v.expectedType = oldExpected }()
	v.expectedType = declaredType

	inferredType := "unknown"
	if declaredType != "" {
		inferredType = declaredType
	} else if ctx.Expr() != nil {
		inferredType = types.InferExprType(ctx.Expr(), v.CurrentScope)
	}

	v.CurrentScope.Define(ctx.ID().GetText(), varName, symbols.KindVar, inferredType)

	if ctx.Expr() != nil {
		exprGoCode := ctx.Expr().Accept(v).(string)
		exprType := types.InferExprType(ctx.Expr(), v.CurrentScope)
		if declaredType != "" {
			exprGoCode = v.applyImplicitPromotion(exprGoCode, exprType, declaredType)
		}

		if strings.Contains(exprGoCode, "?!") || strings.HasPrefix(exprGoCode, "__MYGO_TRY_UNWRAP__") {
			var cleanCall, blockCode string
			if strings.HasPrefix(exprGoCode, "__MYGO_TRY_UNWRAP__") {
				parts := strings.Split(exprGoCode, "__BLOCK__")
				if len(parts) == 2 {
					cleanCall = strings.TrimPrefix(parts[0], "__MYGO_TRY_UNWRAP__")
					blockCode = parts[1]
				} else {
					cleanCall = exprGoCode // fallback
					blockCode = "return err"
				}
			} else {
				cleanCall = strings.ReplaceAll(exprGoCode, "?!", "")
				blockCode = "return err"
			}

			// Clean up block code: remove outer braces if they exist to check content,
			// but we need them for the if body. `VisitBlock` returns braces.
			// `blockCode` is expected to be `{ ... }` or single statement.

			// If blockCode is single statement like "return err", wrap it.
			if !strings.HasPrefix(strings.TrimSpace(blockCode), "{") {
				blockCode = fmt.Sprintf("{\n\t%s\n}", blockCode)
			}

			if declaredType == "" {
				return fmt.Sprintf("_tmp, err := %s\nif err != nil %s\n%s := _tmp", cleanCall, blockCode, varName)
			}
			return fmt.Sprintf("_tmp, err := %s\nif err != nil %s\nvar %s %s = _tmp", cleanCall, blockCode, varName, declaredType)
		}
		_, isArrayLiteral := ctx.Expr().(*ast.ArrayLiteralExprContext)
		if isArrayLiteral {
			if declaredType == "" {
				return fmt.Sprintf("%s := %s%s", varName, types.SniffType(ctx.Expr()), exprGoCode)
			}
			return fmt.Sprintf("var %s %s = %s%s", varName, declaredType, declaredType, exprGoCode)
		}
		if declaredType == "" {
			return fmt.Sprintf("%s := %s", varName, exprGoCode)
		}
		return fmt.Sprintf("var %s %s = %s", varName, declaredType, exprGoCode)
	}
	return fmt.Sprintf("var %s %s", varName, declaredType)
}

func (v *MyGoTranspiler) VisitTupleLetDecl(ctx *ast.TupleLetDeclContext) interface{} {
	var ids []string
	for _, idCtx := range ctx.AllID() {
		goName := types.FormatVisibility(idCtx.GetText(), ctx.Modifier())
		ids = append(ids, goName)
		v.CurrentScope.Define(idCtx.GetText(), goName, symbols.KindVar, "unknown")
	}
	exprStr := ctx.Expr().Accept(v).(string)
	if strings.Contains(exprStr, "?!") {
		cleanCall := strings.ReplaceAll(exprStr, "?!", "")
		return fmt.Sprintf("%s, err := %s\nif err != nil {\n\treturn err\n}", strings.Join(ids, ", "), cleanCall)
	}
	return fmt.Sprintf("%s := %s", strings.Join(ids, ", "), exprStr)
}

func (v *MyGoTranspiler) VisitConstDecl(ctx *ast.ConstDeclContext) interface{} {
	varName := types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier())

	declaredType := ""
	if ctx.TypeType() != nil {
		declaredType = v.resolveType(ctx.TypeType())
	}

	inferredType := "unknown"
	if declaredType != "" {
		inferredType = declaredType
	} else if ctx.Expr() != nil {
		inferredType = types.InferExprType(ctx.Expr(), v.CurrentScope)
	}

	v.CurrentScope.Define(ctx.ID().GetText(), varName, symbols.KindVar, inferredType)

	return fmt.Sprintf("const %s %s = %s", varName, declaredType, ctx.Expr().Accept(v).(string))
}
