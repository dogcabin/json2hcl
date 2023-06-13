package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	cty_json "github.com/zclconf/go-cty/cty/json"
)

type Variable struct {
	Description string
	Type        interface{}
}

func convert(value string) error {
	m := map[string]Variable{}
	err := json.Unmarshal([]byte(value), &m)
	if err != nil {
		return err
	}
	return parseMap(m)
}

func parseMap(aMap map[string]Variable) error {
	for key, val := range aMap {
		var wrappedOutput cty_json.SimpleJSONValue
		valType, valTypeErr := json.Marshal(val.Type)
		if valTypeErr != nil {
			return valTypeErr
		}

		err := json.Unmarshal([]byte(valType), &wrappedOutput)
		if err != nil {
			return err
		}
		f := hclwrite.NewEmptyFile()
		rootBody := f.Body()
		varBlock := rootBody.AppendNewBlock("variable", []string{key})
		varBody := varBlock.Body()
		varBody.SetAttributeRaw(
			"type",
			typeExprTokens(wrappedOutput.Type()),
		)
		varBody.SetAttributeValue(
			"description",
			cty.StringVal(val.Description),
		)
		fmt.Printf("%s\n", f.Bytes())
	}
	return nil
}

func typeExprTokens(ty cty.Type) hclwrite.Tokens {
	switch ty {
	case cty.String:
		return hclwrite.TokensForIdentifier("string")
	case cty.Bool:
		return hclwrite.TokensForIdentifier("bool")
	case cty.Number:
		return hclwrite.TokensForIdentifier("number")
	case cty.DynamicPseudoType:
		return hclwrite.TokensForIdentifier("any")
	}

	if ty.IsCollectionType() {
		etyTokens := typeExprTokens(ty.ElementType())
		switch {
		case ty.IsListType():
			return hclwrite.TokensForFunctionCall("list", etyTokens)
		case ty.IsSetType():
			return hclwrite.TokensForFunctionCall("set", etyTokens)
		case ty.IsMapType():
			return hclwrite.TokensForFunctionCall("map", etyTokens)
		default:
			// Should never happen because the above is exhaustive
			panic("unsupported collection type")
		}
	}

	if ty.IsObjectType() {
		atys := ty.AttributeTypes()
		names := make([]string, 0, len(atys))
		for name := range atys {
			names = append(names, name)
		}
		sort.Strings(names)

		items := make([]hclwrite.ObjectAttrTokens, len(names))
		for i, name := range names {
			items[i] = hclwrite.ObjectAttrTokens{
				Name:  hclwrite.TokensForIdentifier(name),
				Value: typeExprTokens(atys[name]),
			}
		}

		return hclwrite.TokensForFunctionCall("object", hclwrite.TokensForObject(items))
	}

	if ty.IsTupleType() {
		etys := ty.TupleElementTypes()
		items := make([]hclwrite.Tokens, len(etys))
		for i, ety := range etys {
			items[i] = typeExprTokens(ety)
		}
		return hclwrite.TokensForFunctionCall("tuple", hclwrite.TokensForTuple(items))
	}

	panic(fmt.Errorf("unsupported type %#v", ty))
}
