// generated by stringer -type=TokenType; DO NOT EDIT

package c6

import "fmt"

const _TokenType_name = "T_SPACET_COMMENT_LINET_COMMENT_BLOCKT_SEMICOLONT_COMMAT_ID_SELECTORT_CLASS_SELECTORT_TAGNAME_SELECTORT_UNIVERSAL_SELECTORT_PARENT_SELECTORT_CHILD_SELECTORT_PSEUDO_SELECTORT_AND_SELECTORT_ADJACENT_SELECTORT_BRACE_STARTT_LANG_CODET_ATTRIBUTE_STARTT_ATTRIBUTE_NAMET_ATTRIBUTE_ENDT_EQUALT_CONTAINST_BRACE_ENDT_VARIABLET_IMPORTT_CHARSETT_QQ_STRINGT_Q_STRINGT_UNQUOTE_STRINGT_PAREN_STARTT_PAREN_ENDT_CONSTANTT_INTEGERT_FLOATT_UNIT_PXT_UNIT_PTT_UNIT_EMT_PROPERTY_NAMET_PROPERTY_VALUET_HEX_COLORT_COLONT_EXPANSION_STARTT_EXPANSION_END"

var _TokenType_index = [...]uint16{0, 7, 21, 36, 47, 54, 67, 83, 101, 121, 138, 154, 171, 185, 204, 217, 228, 245, 261, 276, 283, 293, 304, 314, 322, 331, 342, 352, 368, 381, 392, 402, 411, 418, 427, 436, 445, 460, 476, 487, 494, 511, 526}

func (i TokenType) String() string {
	if i < 0 || i+1 >= TokenType(len(_TokenType_index)) {
		return fmt.Sprintf("TokenType(%d)", i)
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
