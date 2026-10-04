// Review-only shape. This is not a production manifest loader.
package strategyshadow

type FamilyShadow struct { permitted bool }
func Fixture() FamilyShadow { return FamilyShadow{permitted: true} }
func (s FamilyShadow) Permitted() bool { return s.permitted }
