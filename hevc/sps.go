package hevc

import (
	"bytes"
	"fmt"

	"github.com/Eyevinn/mp4ff/avc"
	"github.com/Eyevinn/mp4ff/bits"
)

// SPS - HEVC SPS parameters
// ISO/IEC 23008-2 Sec. 7.3.2.2
type SPS struct {
	VpsID                                byte
	MaxSubLayersMinus1                   byte
	TemporalIDNestingFlag                bool
	ProfileTierLevel                     ProfileTierLevel
	SpsID                                byte
	ChromaFormatIDC                      byte
	SeparateColourPlaneFlag              bool
	ConformanceWindowFlag                bool
	PicWidthInLumaSamples                uint32
	PicHeightInLumaSamples               uint32
	ConformanceWindow                    ConformanceWindow
	BitDepthLumaMinus8                   byte
	BitDepthChromaMinus8                 byte
	Log2MaxPicOrderCntLsbMinus4          byte
	SubLayerOrderingInfoPresentFlag      bool
	SubLayeringOrderingInfos             []SubLayerOrderingInfo
	Log2MinLumaCodingBlockSizeMinus3     byte
	Log2DiffMaxMinLumaCodingBlockSize    byte
	Log2MinLumaTransformBlockSizeMinus2  byte
	Log2DiffMaxMinLumaTransformBlockSize byte
	MaxTransformHierarchyDepthInter      byte
	MaxTransformHierarchyDepthIntra      byte
	ScalingListEnabledFlag               bool
	ScalingListDataPresentFlag           bool
	AmpEnabledFlag                       bool
	SampleAdaptiveOffsetEnabledFlag      bool
	PCMEnabledFlag                       bool
	PcmSampleBitDepthLumaMinus1          byte
	PcmSampleBitDepthChromaMinus1        byte
	Log2MinPcmLumaCodingBlockSize        uint16
	Log2DiffMaxMinPcmLumaCodingBlockSize uint16
	PcmLoopFilterDisabledFlag            bool
	NumShortTermRefPicSets               byte
	ShortTermRefPicSets                  []ShortTermRPS
	LongTermRefPicsPresentFlag           bool
	SpsTemporalMvpEnabledFlag            bool
	StrongIntraSmoothingEnabledFlag      bool
	VUIParametersPresentFlag             bool
	VUI                                  *VUIParameters
}

// ProfileTierLevel according to ISO/IEC 23008-2 Section 7.3.3
type ProfileTierLevel struct {
	GeneralProfileSpace              byte
	GeneralTierFlag                  bool
	GeneralProfileIDC                byte
	GeneralProfileCompatibilityFlags uint32
	GeneralConstraintIndicatorFlags  uint64 // 48 bits
	GeneralProgressiveSourceFlag     bool
	GeneralInterlacedSourceFlag      bool
	GeneralNonPackedConstraintFlag   bool
	GeneralFrameOnlyConstraintFlag   bool
	// 43 + 1 bits of info
	GeneralLevelIDC byte
	// Sublayer stuff

}

// ConformanceWindow according to ISO/IEC 23008-2
type ConformanceWindow struct {
	LeftOffset   uint32
	RightOffset  uint32
	TopOffset    uint32
	BottomOffset uint32
}

// SubLayerOrderingInfo according to ISO/IEC 23008-2
type SubLayerOrderingInfo struct {
	MaxDecPicBufferingMinus1 byte
	MaxNumReorderPics        byte
	MaxLatencyIncreasePlus1  byte
}

// VUIParameters - Visual Usability Information as defined in Section E.2
type VUIParameters struct {
	SampleAspectRatioWidth         uint
	SampleAspectRatioHeight        uint
	OverscanInfoPresentFlag        bool
	OverscanAppropriateFlag        bool
	VideoSignalTypePresentFlag     bool
	VideoFormat                    byte
	VideoFullRangeFlag             bool
	ColourDescriptionFlag          bool
	ColourPrimaries                byte
	TransferCharacteristics        byte
	MatrixCoefficients             byte
	ChromaLocInfoPresentFlag       bool
	ChromaSampleLocTypeTopField    uint
	ChromaSampleLocTypeBottomField uint
	NeutralChromaIndicationFlag    bool
	FieldSeqFlag                   bool
	FrameFieldInfoPresentFlag      bool
	DefaultDisplayWindowFlag       bool
	DefDispWinLeftOffset           uint
	DefDispWinRightOffset          uint
	DefDispWinTopOffset            uint
	DefDispWinBottomOffset         uint
	TimingInfoPresentFlag          bool
	NumUnitsInTick                 uint
	TimeScale                      uint
	PocProportionalToTimingFlag    bool
	NumTicksPocDiffOneMinus1       uint
	HrdParametersPresentFlag       bool
	BitstreamRestrictionFlag       bool
	BitstreamResctrictions         *BitstreamRestrictions
}

// BitstreamRestrictrictions - optional information
type BitstreamRestrictions struct {
	TilesFixedStructureFlag     bool
	MVOverPicBoundariesFlag     bool
	RestrictedRefsPicsListsFlag bool
	MinSpatialSegmentationIDC   uint
	MaxBytesPerPicDenom         uint
	MaxBitsPerMinCuDenom        uint
	Log2MaxMvLengthHorizontal   uint
	Log2MaxMvLengthVertical     uint
}

// ParseSPSNALUnit parses SPS NAL unit starting with NAL unit header
func ParseSPSNALUnit(data []byte) (*SPS, error) {

	sps := &SPS{}

	rd := bytes.NewReader(data)
	r := bits.NewAccErrEBSPReader(rd)
	// Note! First two bytes are NALU Header

	naluHdrBits := r.Read(16)
	naluType := GetNaluType(byte(naluHdrBits >> 8))
	if naluType != NALU_SPS {
		return nil, fmt.Errorf("NALU type is %s not SPS", naluType)
	}
	sps.VpsID = byte(r.Read(4))
	sps.MaxSubLayersMinus1 = byte(r.Read(3))
	sps.TemporalIDNestingFlag = r.ReadFlag()
	sps.ProfileTierLevel.GeneralProfileSpace = byte(r.Read(2))
	sps.ProfileTierLevel.GeneralTierFlag = r.ReadFlag()
	sps.ProfileTierLevel.GeneralProfileIDC = byte(r.Read(5))
	sps.ProfileTierLevel.GeneralProfileCompatibilityFlags = uint32(r.Read(32))
	sps.ProfileTierLevel.GeneralConstraintIndicatorFlags = uint64(r.Read(48))
	sps.ProfileTierLevel.GeneralLevelIDC = byte(r.Read(8))
	if sps.MaxSubLayersMinus1 != 0 {
		return sps, nil // Cannot parse any further
	}
	sps.SpsID = byte(r.ReadExpGolomb())
	sps.ChromaFormatIDC = byte(r.ReadExpGolomb())
	if sps.ChromaFormatIDC == 3 {
		sps.SeparateColourPlaneFlag = r.ReadFlag()
	}
	sps.PicWidthInLumaSamples = uint32(r.ReadExpGolomb())
	sps.PicHeightInLumaSamples = uint32(r.ReadExpGolomb())
	sps.ConformanceWindowFlag = r.ReadFlag()
	if sps.ConformanceWindowFlag {
		sps.ConformanceWindow = ConformanceWindow{
			LeftOffset:   uint32(r.ReadExpGolomb()),
			RightOffset:  uint32(r.ReadExpGolomb()),
			TopOffset:    uint32(r.ReadExpGolomb()),
			BottomOffset: uint32(r.ReadExpGolomb()),
		}
	}
	sps.BitDepthLumaMinus8 = byte(r.ReadExpGolomb())
	sps.BitDepthChromaMinus8 = byte(r.ReadExpGolomb())
	sps.Log2MaxPicOrderCntLsbMinus4 = byte(r.ReadExpGolomb())
	sps.SubLayerOrderingInfoPresentFlag = r.ReadFlag()
	startValue := sps.MaxSubLayersMinus1
	if sps.SubLayerOrderingInfoPresentFlag {
		startValue = 0
	}
	for i := startValue; i <= sps.MaxSubLayersMinus1; i++ {
		sps.SubLayeringOrderingInfos = append(
			sps.SubLayeringOrderingInfos,
			SubLayerOrderingInfo{
				MaxDecPicBufferingMinus1: byte(r.ReadExpGolomb()),
				MaxNumReorderPics:        byte(r.ReadExpGolomb()),
				MaxLatencyIncreasePlus1:  byte(r.ReadExpGolomb()),
			})
	}
	sps.Log2MinLumaCodingBlockSizeMinus3 = byte(r.ReadExpGolomb())
	sps.Log2DiffMaxMinLumaCodingBlockSize = byte(r.ReadExpGolomb())
	sps.Log2MinLumaTransformBlockSizeMinus2 = byte(r.ReadExpGolomb())
	sps.Log2DiffMaxMinLumaTransformBlockSize = byte(r.ReadExpGolomb())
	sps.MaxTransformHierarchyDepthInter = byte(r.ReadExpGolomb())
	sps.MaxTransformHierarchyDepthIntra = byte(r.ReadExpGolomb())
	sps.ScalingListEnabledFlag = r.ReadFlag()
	if sps.ScalingListEnabledFlag {
		sps.ScalingListDataPresentFlag = r.ReadFlag()
		if sps.ScalingListDataPresentFlag {
			readPastScalingListData(r)
		}
	}
	sps.AmpEnabledFlag = r.ReadFlag()
	sps.SampleAdaptiveOffsetEnabledFlag = r.ReadFlag()
	sps.PCMEnabledFlag = r.ReadFlag()
	if sps.PCMEnabledFlag {
		sps.PcmSampleBitDepthLumaMinus1 = byte(r.Read(4))
		sps.PcmSampleBitDepthChromaMinus1 = byte(r.Read(4))
		sps.Log2MinPcmLumaCodingBlockSize = uint16(r.ReadExpGolomb())
		sps.Log2DiffMaxMinPcmLumaCodingBlockSize = uint16(r.ReadExpGolomb())
		sps.PcmLoopFilterDisabledFlag = r.ReadFlag()
	}
	sps.NumShortTermRefPicSets = byte(r.ReadExpGolomb())
	for idx := byte(0); idx < sps.NumShortTermRefPicSets; idx++ {
		sps.ShortTermRefPicSets = append(sps.ShortTermRefPicSets, parseShortTermRPS(r, idx, sps.NumShortTermRefPicSets, sps))
		if r.AccError() != nil { // Don't continue if we have an issue
			return sps, r.AccError()
		}
	}

	sps.LongTermRefPicsPresentFlag = r.ReadFlag()
	if sps.LongTermRefPicsPresentFlag {
		// Get passed this without storing the information
		numLongTermRefPics := r.ReadExpGolomb()
		for i := uint(0); i < numLongTermRefPics; i++ {
			/* ltRefPicPocLsbSps[i] */ _ = r.Read(int(sps.Log2MaxPicOrderCntLsbMinus4 + 4))
			/* usedByCurrPicLtSps = */ _ = r.ReadFlag()
		}
	}
	sps.SpsTemporalMvpEnabledFlag = r.ReadFlag()
	sps.StrongIntraSmoothingEnabledFlag = r.ReadFlag()
	sps.VUIParametersPresentFlag = r.ReadFlag()
	if sps.VUIParametersPresentFlag {
		sps.VUI = parseVUI(r)
	}

	return sps, r.AccError()
}

// ImageSize - calculated width and height using ConformanceWindow
func (s *SPS) ImageSize() (width, height uint32) {
	encWidth, encHeight := s.PicWidthInLumaSamples, s.PicHeightInLumaSamples
	var subWidthC, subHeightC uint32 = 1, 1
	switch s.ChromaFormatIDC {
	case 1: // 4:2:0
		subWidthC, subHeightC = 2, 2
	case 2: // 4:2:2
		subWidthC = 2
	}
	width = encWidth - (s.ConformanceWindow.LeftOffset+s.ConformanceWindow.RightOffset)*subWidthC
	height = encHeight - (s.ConformanceWindow.TopOffset+s.ConformanceWindow.BottomOffset)*subHeightC
	return width, height
}

// parseVUI - parse VUI (Visual Usability Information)
// if parseVUIBeyondAspectRatio is false, stop after AspectRatio has been parsed
func parseVUI(r *bits.AccErrEBSPReader) *VUIParameters {
	vui := &VUIParameters{}
	aspectRatioInfoPresentFlag := r.ReadFlag()
	if aspectRatioInfoPresentFlag {
		aspectRatioIDC := r.Read(8)
		if aspectRatioIDC == avc.ExtendedSAR {
			vui.SampleAspectRatioWidth = r.Read(16)
			vui.SampleAspectRatioHeight = r.Read(16)
		} else {
			var err error
			vui.SampleAspectRatioWidth, vui.SampleAspectRatioHeight, err = avc.GetSARfromIDC(aspectRatioIDC)
			if err != nil {
				r.SetError(fmt.Errorf("GetSARFromIDC: %w", err))
			}
		}
	}
	vui.OverscanInfoPresentFlag = r.ReadFlag()
	if vui.OverscanInfoPresentFlag {
		vui.OverscanAppropriateFlag = r.ReadFlag()
	}
	vui.VideoSignalTypePresentFlag = r.ReadFlag()
	if vui.VideoSignalTypePresentFlag {
		vui.VideoFormat = byte(r.Read(3))
		vui.VideoFullRangeFlag = r.ReadFlag()
		vui.ColourDescriptionFlag = r.ReadFlag()
		if vui.ColourDescriptionFlag {
			vui.ColourPrimaries = byte(r.Read(8))
			vui.TransferCharacteristics = byte(r.Read(8))
			vui.MatrixCoefficients = byte(r.Read(8))
		}
	}
	vui.ChromaLocInfoPresentFlag = r.ReadFlag()
	if vui.ChromaLocInfoPresentFlag {
		vui.ChromaSampleLocTypeTopField = r.ReadExpGolomb()
		vui.ChromaSampleLocTypeBottomField = r.ReadExpGolomb()
	}
	vui.NeutralChromaIndicationFlag = r.ReadFlag()
	vui.FieldSeqFlag = r.ReadFlag()
	vui.FrameFieldInfoPresentFlag = r.ReadFlag()
	vui.DefaultDisplayWindowFlag = r.ReadFlag()
	if vui.DefaultDisplayWindowFlag {
		vui.DefDispWinLeftOffset = r.ReadExpGolomb()
		vui.DefDispWinRightOffset = r.ReadExpGolomb()
		vui.DefDispWinTopOffset = r.ReadExpGolomb()
		vui.DefDispWinBottomOffset = r.ReadExpGolomb()
	}
	vui.TimingInfoPresentFlag = r.ReadFlag()
	if vui.TimingInfoPresentFlag {
		vui.NumUnitsInTick = r.Read(32)
		vui.TimeScale = r.Read(32)
		vui.PocProportionalToTimingFlag = r.ReadFlag()
		if vui.PocProportionalToTimingFlag {
			vui.NumTicksPocDiffOneMinus1 = r.ReadExpGolomb()
		}
		vui.HrdParametersPresentFlag = r.ReadFlag()
		if vui.HrdParametersPresentFlag {
			//TODO.  Add HrdParameters parsing according to E.3.2 if needed
			return vui
		}
	}
	vui.BitstreamRestrictionFlag = r.ReadFlag()
	if vui.BitstreamRestrictionFlag {
		vui.BitstreamResctrictions = parseBitstreamRestrictions(r)
	}

	return vui
}

func parseBitstreamRestrictions(r *bits.AccErrEBSPReader) *BitstreamRestrictions {
	br := BitstreamRestrictions{}
	br.TilesFixedStructureFlag = r.ReadFlag()
	br.MVOverPicBoundariesFlag = r.ReadFlag()
	br.RestrictedRefsPicsListsFlag = r.ReadFlag()
	br.MinSpatialSegmentationIDC = r.ReadExpGolomb()
	br.MaxBytesPerPicDenom = r.ReadExpGolomb()
	br.MaxBitsPerMinCuDenom = r.ReadExpGolomb()
	br.Log2MaxMvLengthHorizontal = r.ReadExpGolomb()
	br.Log2MaxMvLengthVertical = r.ReadExpGolomb()
	return &br
}

// ShortTermRPS - Short term Reference Picture Set
type ShortTermRPS struct {
	// Delta Picture Order Count
	DeltaPocS0      []uint32
	DeltaPocS1      []uint32
	UsedByCurrPicS0 []bool
	UsedByCurrPicS1 []bool
	NumNegativePics byte
	NumPositivePics byte
	NumDeltaPocs    byte
}

const maxSTRefPics = 16

// parseShortTermRPS - short-term reference pictures with syntax from 7.3.7.
// Focus is on reading/parsing beyond this structure in SPS (and possibly in slice header)
func parseShortTermRPS(r *bits.AccErrEBSPReader, idx, numSTRefPicSets byte, sps *SPS) ShortTermRPS {
	stps := ShortTermRPS{}

	interRPSPredFlag := false

	if idx > 0 {
		interRPSPredFlag = r.ReadFlag()
	}
	if interRPSPredFlag {
		deltaIdx := byte(1)
		if idx == numSTRefPicSets { // Slice header
			deltaIdx = byte(r.ReadExpGolomb() + 1)
			// parse delta_idx_minus1
		}
		if deltaIdx > idx {
			r.SetError(fmt.Errorf("deltaIdx > idx in parseShortTermRPS"))
		}
		/* deltaRpsSign */ _ = r.Read(1)
		/* absDeltaRpsMinus1*/ _ = r.ReadExpGolomb()
		//deltaRps := (1 - (deltaRpsSign << 1)) * (absDeltaRpsMinus1 + 1)
		refIdx := idx - deltaIdx
		numDeltaPocs := sps.ShortTermRefPicSets[refIdx].NumDeltaPocs
		for j := byte(0); j < numDeltaPocs; j++ {
			usedByCurrPicFlag := r.ReadFlag()
			useDeltaFlag := false
			if !usedByCurrPicFlag {
				useDeltaFlag = r.ReadFlag()
			}
			if usedByCurrPicFlag || useDeltaFlag {
				stps.NumDeltaPocs++
			}
		}
	} else {
		stps.NumNegativePics = byte(r.ReadExpGolomb())
		stps.NumPositivePics = byte(r.ReadExpGolomb())
		if stps.NumNegativePics > maxSTRefPics || stps.NumPositivePics > maxSTRefPics {
			r.SetError(fmt.Errorf("more than %d short term reference pictures", maxSTRefPics))
			return stps
		}
		stps.NumDeltaPocs = stps.NumNegativePics + stps.NumPositivePics
		stps.DeltaPocS0 = make([]uint32, stps.NumNegativePics)
		stps.UsedByCurrPicS0 = make([]bool, stps.NumNegativePics)
		for i := byte(0); i < stps.NumNegativePics; i++ {
			stps.DeltaPocS0[i] = uint32(r.ReadExpGolomb() + 1)
			stps.UsedByCurrPicS0[i] = r.ReadFlag()
		}
		stps.DeltaPocS1 = make([]uint32, stps.NumPositivePics)
		stps.UsedByCurrPicS1 = make([]bool, stps.NumPositivePics)
		for i := byte(0); i < stps.NumPositivePics; i++ {
			stps.DeltaPocS1[i] = uint32(r.ReadExpGolomb() + 1)
			stps.UsedByCurrPicS1[i] = r.ReadFlag()
		}
	}

	return stps
}

// readPastScalingListData - read and parse all bits of scaling list, without storing values
func readPastScalingListData(r *bits.AccErrEBSPReader) {
	for sizeId := 0; sizeId < 4; sizeId++ {
		nrMatrixIds := 6
		if sizeId == 3 {
			nrMatrixIds = 2
		}
		for matrixId := 0; matrixId < nrMatrixIds; matrixId++ {
			flag := r.ReadFlag() // scaling_list_pred_mode_flag[sizeId][matrixId]
			if !flag {
				_ = r.ReadExpGolomb() // scaling_list_pred_matrix_id_delta[sizeId][matrixId]
			} else {
				// nextCoef = 8;
				coefNum := (1 << (4 + (sizeId << 1)))
				if coefNum > 64 {
					coefNum = 64
				}
				if sizeId > 1 {
					_ = r.ReadExpGolomb // scaling_list_dc_coef_minus8[sizeId − 2][matrixId]
					// nextCoef = scaling_list_dc_coef_minus8[sizeId − 2][matrixId] + 8
				}
				for i := 0; i < coefNum; i++ {
					_ = r.ReadExpGolomb() // scaling_list_delta_coef
					// nextCoef = ( nextCoef + scaling_list_delta_coef + 256 ) % 256
					// ScalingList[sizeId][matrixId][i] = nextCoef
				}
			}
		}
	}
}
