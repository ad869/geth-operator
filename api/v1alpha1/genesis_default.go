package v1alpha1

func (g *Genesis) Default() {
	if g.NetworkID == 0 {
		g.NetworkID = DefaultNetworkID
	}

	if g.ChainID == 0 {
		g.ChainID = DefaultChainID
	}

	if g.Istanbul.Ceil2Nby3Block == 0 {
		g.Istanbul.Ceil2Nby3Block = DefaultIstanbulCeil2Nby3Block
	}

	if g.Istanbul.Epoch == 0 {
		g.Istanbul.Epoch = DefaultIstanbulEpoch
	}

	if g.Istanbul.Policy == 0 {
		g.Istanbul.Policy = DefaultIstanbulPolicy
	}

	if g.Istanbul.TestQBFTBlock == 0 {
		g.Istanbul.TestQBFTBlock = DefaultIstanbulTestQBFTBlock
	}

	// if g.QBFT.EpochLength == 0 {
	// 	g.QBFT.EpochLength = DefaultQBFTEpochLength
	// }

	// if g.QBFT.BlockPeriodSeconds == 0 {
	// 	g.QBFT.BlockPeriodSeconds = DefaultQBFTBlockPeriodSeconds
	// }

	// if g.QBFT.RequestTimeoutSeconds == 0 {
	// 	g.QBFT.RequestTimeoutSeconds = DefaultQBFTRequestTimeoutSeconds
	// }
}
