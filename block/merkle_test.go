package block

import "testing"

func TestConstructMerkleRoot(t *testing.T) {
	var testcases = []struct {
		txids []string
		merkleRoot string
	} {
		{
			[]string {
				"b1fea52486ce0c62bb442b530a3f0132b826c74e473d1f2c220bfa78111c5082",
				"f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16",
			},
			"7dac2c5666815c17a3b36427de37bb9d2e2c5ccec3f8633eb91a4205cb4c10ff",
		},
		{
			[]string {
				"8347cee4a1cb5ad1bb0d92e86e6612dbf6cfc7649c9964f210d4069b426e720a",
				"a16f3ce4dd5deb92d98ef5cf8afeaf0775ebca408f708b2146c4fb42b41e14be",
			},
			"ed92b1db0b3e998c0a4351ee3f825fd5ac6571ce50c050b4b45df015092a6c36",
		},
		{
			[]string {
				"09e5c4a5a089928bbe368cd0f2b09abafb3ebf328cd0d262d06ec35bdda1077f",
				"591e91f809d716912ca1d4a9295e70c3e78bab077683f79350f101da64588073",
			},
			"2f0f017f1991a1393798ff851bfc02ce7ba3f5e066815ed3104afb4bd3a0c230",
		},
		{
			[]string{
				"8c14f0db3df150123e6f3dbbf30f8b955a8249b62ac1d1ff16284aefa3d06d87",
				"fff2525b8931402dd09222c50775608f75787bd2b87e56995a7bdd30f79702c4",
				"6359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4",
				"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d",
			},
			"f3e94742aca4b5ef85488dc37c06c3282295ffec960994b2c0d5ac2a25a95766",
		},
		{
			[]string {
				"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			},
			"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		},
		{
			[]string {
				"ef1d870d24c85b89d92ad50f4631026f585d6a34e972eaf427475e5d60acf3a3",
				"f9fc751cb7dc372406a9f8d738d5e6f8f63bab71986a39cf36ee70ee17036d07",
				"db60fb93d736894ed0b86cb92548920a3fe8310dd19b0da7ad97e48725e1e12e",
				"220ebc64e21abece964927322cba69180ed853bb187fbc6923bac7d010b9d87a",
				"71b3dbaca67e9f9189dad3617138c19725ab541ef0b49c05a94913e9f28e3f4e",
				"fe305e1ed08212d76161d853222048eea1f34af42ea0e197896a269fbf8dc2e0",
				"21d2eb195736af2a40d42107e6abd59c97eb6cffd4a5a7a7709e86590ae61987",
				"dd1fd2a6fc16404faf339881a90adbde7f4f728691ac62e8f168809cdfae1053",
				"74d681e0e03bafa802c8aa084379aa98d9fcd632ddc2ed9782b586ec87451f20",
			},
			"2fda58e5959b0ee53c5253da9b9f3c0c739422ae04946966991cf55895287552",
		},
		{
			[]string{
				"e8657d6b1f9a7ddd4fd7163876f322790ee78fe09c70bd34abcc1de8936789d1",
				"39b45f60bb7be2702994d60ed18fc4802954601d6748f9da3d6e36c7384bcf65",
				"eced1e416d3e500ba182d196caad574af421ac89f89e8c975255e345748c5531",
				"8a055e75741af14231a9ab1bfd01cc2752ab93c811835d5c2ee1308a185f6f43",
				"8fb08cefec5b5b9feb64d81d3c6a7570ac33a7a4f3f8e09f337e95b53aa6058b",
				"588f6c03789d47ca1b6d6db767480f9e437b81307ab84dd25fc4f5b63c486da6",
				"b6524aa95129f721d959f1da789e671083225bacd7730c037e4f9a41695e3d37",
				"72ae7de757d86e7d72e56d336ada2d7c03e12f080469997de61ea458f8842965",
				"8d799d46e99bde3f06421d052050e33ea6a9b6323e1a51b2012e71c7a48d57e3",
				"3fc0e866e2c80665b9bcf4ce3b150b09ef5f4507391d2f9d852d8220085b4bb8",
				"7692856814b809d2e34a09315312447d457b8e15cae491532198e83f72c01ef2",
				"ea22b279c04ece713d95e2e6d66e966767908790e19a96acc58c6b9d9ef8d5de",
				"41ee1ab6cd28f0f1e3abbea517d0e111d7be63bef96e00de356982ec294e5559",
				"801e322b513eb3bf50bbf58b3223c14fc684401f6c232d8841ab194bef0bed7b",
				"0ef6d67081b0782dcbe4590f1a926442d62ee09375ecef5cb18e162e085931d7",
				"454cd0f656fac460b9c3bf404187ada1cf594068895b145d80135ef21f1bca4e",
				"278a3cf741317e9c5efa5e3508f1fcc8bac029a720b066d9f040029c5fd3d19a",
				"e347830cffdf577136a62717acd68343a8cc6403c7f0d71ab66b612adeefe747",
				"bcb690d60e306914919c7b4a76d1698422190a1635691f64ec25f9d8e39ea5c4",
				"bb0a269a38d80626c2c7c896076ea5025dc91e8aab2e5b881dd295e0b8396a6e",
				"90ae8b227cdc6dd1bc3c69f67ec4369c38415a054784219f3d811089c8643c39",
				"b98eba44cf6f69604c0b32c0bd4de0ee2456b92d8396764019a7f922e967c012",
				"2c457a98d4239c048a75792667c4c7cd93439250fb582b1b6c28f4013323f68e",
				"626ea516a0aaf0a57c58221d2941a522a92c725d3e8cc74f4e6adc0f2abaed67",
				"d5e37c9b1d1cc344454c3752a3afd8e579e6136cd7dc146cce29a992be1b713e",
				"3210e68b92bc21becd0a3b4f8fc5da0e988976323908b2bf324696642c4cd1d9",
				"eea4d9b03d30fee1b8225e5857ac01437e47fb819fec7abbddfe4a407f932f0e",
				"d1072a19e321ab9d44098fd725f12ba7b918565b24807f04ccd534cd16fdc5dd",
				"34e7cd30a58283a00e3c55e6cf0f995b30cd01c393eccc33efb4179c4c044ff1",
				"7c7e325836580561dc8234352c2f9261872bfc437bef28403b3d40dc50c9f1a5",
				"a50f12e84db067a6339388188c309d09a399335ca899496f1b2ea3fb110645e8",
				"1110ca16267875f2dabdc426abfeab274883fbcf849bd8636b44c70b1a77ae34",
				"5d1435646eea2d3a8a57e207ff8b7479e633712f1dde716f159bbe6277b4eef8",
				"0cc3ffa7985f9ef8db120070cdbc7e339c3120ab3a9b25ab9205094e0de25220",
				"2eb9c05d85b1bef23ce778989cb897a03d4c48de556761a1a7bc44440487f0da",
				"a0a7e9a2534afa9096bb89cf42beee1b92ccdc36a042c0ff92db1cb4a7d87b50",
				"3ccbdf3c127c8e48e9d77022aa813c9290249d306dfb24167fd9de1b273f2c67",
				"35499faf8d0c8a39e5ecf3fdb97acd5638402fece737822987d70672fef3073a",
				"d242ba2f77a38e02b4c605f963cefad81fe8c1b3c44e433345d661e4af3ceb15",
				"efd88a572db2773602ad0ee4235b8cf20c1cacb82f146768ae830eca647e371b",
				"0c3f49ad269c480775c738e4b3e17162fb5850081da37505e770ecea550fdade",
				"a2f85652f41a35c0aaae886f77244ed8265071dd2eb1f4a8e03a3709c84a7bf1",
				"250f7a019302df031c999782dd225864308465843fdb5b51131dab0b0fcb8d53",
				"b44ce23ecb22404f248b2a7a151e0de8a92c5032534f6280e64bb82083767bb1",
				"0bcf230c1ae82b8e2893bfc383b1a9a8739b9c37877c648525b9e5322469dcc0",
				"936d429096f327a987222603d29f96fbd41562d908823e7074d376198872c018",
				"c66900da0f7394a7d0a6411b1b4c4fcfb38f56edf415164365aa7ec99ea80fbf",
			},
			"1b2d3ecb6f2ae490d300396d71939159b2692e73c3c5bb41c0f99d183ab369ba",
		},
	}

	for _, oneCase := range testcases {
		var got *MerkleNode
		var err error
		if got,err = ConstructMerkleRoot(oneCase.txids); err != nil {
			t.Error(err)
			return
		}
		if got.Value != oneCase.merkleRoot {
			t.Error("construct merkle root error")
			t.Error("want: ", oneCase.merkleRoot)
			t.Error("got: ", got.Value)
			return
		}
	}
}
