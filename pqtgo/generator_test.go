package pqtgo_test

//
//func assertGoCode(t *testing.T, s1, s2, msg string, com ...interface{}) {
//	s1 = fmt.Sprintf("%s", s1)
//	s2 = fmt.Sprintf("%s", s2)
//	tmp1 := strings.Split(s1, "\n")
//	tmp2 := strings.Split(s2, "\n")
//	if s1 != s2 {
//		b := bytes.NewBuffer(nil)
//		for _, diff := range difflib.Diff(tmp1, tmp2) {
//			p := strings.Replace(diff.Payload, "\t", "\\t", -1)
//			switch diff.Delta {
//			case difflib.Common:
//				fmt.Fprintf(b, "%s %s\n", diff.Delta.String(), p)
//			case difflib.LeftOnly:
//				fmt.Fprintf(b, "\033[31m%s %s\033[39m\n", diff.Delta.String(), p)
//			case difflib.RightOnly:
//				fmt.Fprintf(b, "\033[32m%s %s\033[39m\n", diff.Delta.String(), p)
//			}
//		}
//		t.Errorf(b.String())
//	}
//}
