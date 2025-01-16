package cmd

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"qcow2-analyze/pkg/consts"
	"qcow2-analyze/pkg/structs"
	"qcow2-analyze/pkg/utils"
)

var QCow2FilePath string
var Hexadecimal bool
var Verbose bool
var Output string

func init() {
	rootCmd.PersistentFlags().StringVarP(&QCow2FilePath, "file", "f", "", "qcow2 file path")
	rootCmd.PersistentFlags().BoolVarP(&Hexadecimal, "hex", "H", false, "hexadecimal output, not supported now")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&Output, "output", "o", "raw", "output format, only json now")
}

var rootCmd = &cobra.Command{
	Use:   "qcow2-analyze",
	Short: "analyze a qcow2 file",
	Long:  `analyze a qcow2 file with output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return qcow2Analyze()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func decodeQCowHeader(qcow2HeaderBuf []byte) *structs.QCowHeader {
	var header structs.QCowHeader
	header.Magic = string(qcow2HeaderBuf[0:4])

	version, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[4:8]), 16, 32)
	header.Version = uint32(version)

	backingFileOffset, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[8:16]), 16, 64)
	header.BackingFileOffset = backingFileOffset

	backingFileSize, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[16:20]), 16, 32)
	header.BackingFileSize = uint32(backingFileSize)

	clusterBits, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[20:24]), 16, 32)
	header.ClusterBits = uint32(clusterBits)

	virtualSize, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[24:32]), 16, 64)
	header.VirtualSize = virtualSize

	cryptoMethod, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[32:36]), 16, 32)
	header.CryptMethod = uint32(cryptoMethod)

	l1Size, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[36:40]), 16, 32)
	header.L1Size = uint32(l1Size)

	l1TableOffset, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[40:48]), 16, 64)
	header.L1TableOffset = l1TableOffset

	refcountTableOffset, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[48:56]), 16, 64)
	header.RefcountTableOffset = refcountTableOffset

	refcountTableClusters, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[56:60]), 16, 32)
	header.RefcountTableClusters = uint32(refcountTableClusters)

	numberSnapshots, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[60:64]), 16, 32)
	header.NumberSnapshots = uint32(numberSnapshots)

	snapshotsOffset, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[64:72]), 16, 64)
	header.SnapshotsOffset = snapshotsOffset

	incompatibleFeatures, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[72:80]), 16, 64)
	header.IncompatibleFeatures = incompatibleFeatures

	compatibleFeatures, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[80:88]), 16, 64)
	header.CompatibleFeatures = compatibleFeatures

	autoClearFeatures, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[88:96]), 16, 64)
	header.AutoClearFeatures = autoClearFeatures

	refcountOrder, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[96:100]), 16, 32)
	header.RefcountOrder = uint32(refcountOrder)

	headerLength, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[100:104]), 16, 32)
	header.HeaderLength = uint32(headerLength)

	compressionType, _ := strconv.ParseUint(hex.EncodeToString(qcow2HeaderBuf[104:104]), 16, 8)
	header.CompressionType = uint8(compressionType)

	return &header
}

func validateQCowHeader(qcow2Header *structs.QCowHeader) error {
	if qcow2Header.Magic != consts.QCowMagicString {
		return errors.New(fmt.Sprintf("invalid magic: [%s]", qcow2Header.Magic))
	}

	if qcow2Header.Version != 2 && qcow2Header.Version != 3 {
		return errors.New(fmt.Sprintf("invalid version: [%d]", qcow2Header.Version))
	}

	if qcow2Header.BackingFileSize >= 1024 {
		return errors.New(fmt.Sprintf("invalid backing_file_size: [%d]", qcow2Header.BackingFileSize))
	}

	if qcow2Header.ClusterBits <= 8 {
		return errors.New(fmt.Sprintf("invalid cluster_bits: [%d]", qcow2Header.ClusterBits))
	}

	if qcow2Header.CryptMethod != 0 &&
		qcow2Header.CryptMethod != 1 &&
		qcow2Header.CryptMethod != 2 {
		return errors.New(fmt.Sprintf("invalid crypt_method: [%d]", qcow2Header.CryptMethod))
	}

	return nil
}

func calculateImageEndOffset(qcow2File *os.File, qcow2Header *structs.QCowHeader) uint64 {
	var imageEndOffset uint64

	clusterSize := 1 << qcow2Header.ClusterBits
	/* refcount block entry value in bytes */
	refcountWidth := (1 << qcow2Header.RefcountOrder) / 8
	/* Number of refcount block entry X cluster size */
	wholeDataSize := (clusterSize / refcountWidth) * clusterSize

	/* refcount block */
	var lastRCBOffset uint64
	/* read 512 bytes from file every time */
	var length512Buf = make([]byte, consts.ReadMaxLength)
	/* clusterSize * refcountTableClusters is all bytes store the refcount entry */
	/* refcount entry is 64 bits(8 bytes) */
	for i := 0; i < clusterSize*int(qcow2Header.RefcountTableClusters)/consts.ReadMaxLength; i++ {
		readOffset := qcow2Header.RefcountTableOffset + uint64(i*consts.ReadMaxLength)
		_, _ = qcow2File.ReadAt(length512Buf, int64(readOffset))

		var rcbOffset uint64
		for j := 0; j < consts.ReadMaxLength/consts.RefcountEntryLen; j++ {
			rcbStart := j * consts.RefcountEntryLen
			rcbEnd := (j + 1) * consts.RefcountEntryLen
			if rcbOffset, _ = strconv.ParseUint(hex.EncodeToString(
				length512Buf[rcbStart:rcbEnd]), 16, 64); rcbOffset == 0 {
				break
			}

			imageEndOffset += uint64(wholeDataSize)
			lastRCBOffset = rcbOffset
		}

		if rcbOffset == 0 {
			break
		}
	}

	if lastRCBOffset == 0 {
		imageEndOffset = 0
		return imageEndOffset
	}

	imageEndOffset = imageEndOffset - uint64(wholeDataSize)
	/* refcount block entry stored in a whole cluster, read until 0 */
	for i := 0; i < clusterSize/consts.ReadMaxLength; i++ {
		readOffset := lastRCBOffset + uint64(i*consts.ReadMaxLength)
		_, _ = qcow2File.ReadAt(length512Buf, int64(readOffset))

		var rcbEntryValue uint64
		for j := 0; j < consts.ReadMaxLength/refcountWidth; j++ {
			rcbEntryStart := j * refcountWidth
			rcbEntryEnd := (j + 1) * refcountWidth
			if rcbEntryValue, _ = strconv.ParseUint(hex.EncodeToString(
				length512Buf[rcbEntryStart:rcbEntryEnd]), 16, 8*int(refcountWidth)); rcbEntryValue == 0 {
				break
			}

			imageEndOffset += uint64(clusterSize)
		}

		if rcbEntryValue == 0 {
			break
		}
	}

	return imageEndOffset
}

func qcow2Analyze() error {
	if QCow2FilePath == "" {
		return errors.New("qcow2 file must specified")
	}

	if !utils.FileExist(QCow2FilePath) {
		return errors.New(fmt.Sprintf("qcow2 file [%s] not exist", QCow2FilePath))
	}

	// todo: check qcow2 format

	var err error
	var qcow2File *os.File
	if qcow2File, err = os.Open(QCow2FilePath); err != nil {
		return err
	}
	defer qcow2File.Close()

	var qcow2HeaderBuf = make([]byte, consts.Qcow2HeaderLength)
	if _, err = qcow2File.Read(qcow2HeaderBuf); err != nil {
		return errors.New(fmt.Sprintf("qcow2 header is invalid, please check......"))
	}

	qcow2Header := decodeQCowHeader(qcow2HeaderBuf)
	if err = validateQCowHeader(qcow2Header); err != nil {
		return err
	}

	var backingFile string
	if qcow2Header.BackingFileOffset != 0 && qcow2Header.BackingFileSize > 0 {
		var backingFileBuf = make([]byte, qcow2Header.BackingFileSize)
		if _, err = qcow2File.ReadAt(backingFileBuf, int64(qcow2Header.BackingFileOffset)); err == nil {
			backingFile = string(backingFileBuf)
		}
	}

	var cryptoMethod string
	switch qcow2Header.CryptMethod {
	case 0:
		cryptoMethod = consts.CryptMethodNo
	case 1:
		cryptoMethod = consts.CryptMethodAES
	case 2:
		cryptoMethod = consts.CryptMethodLUKS
	default:
		return errors.New(fmt.Sprintf("unsupported crypt_method: %d", qcow2Header.CryptMethod))
	}

	var clusterSize uint32
	clusterSize = 1 << qcow2Header.ClusterBits

	if qcow2Header.NumberSnapshots != 0 && qcow2Header.SnapshotsOffset > 0 {
		// todo
	}

	imageEndOffset := calculateImageEndOffset(qcow2File, qcow2Header)
	qcowOutput := structs.QCowOutput{
		QCowVersion:    qcow2Header.Version,
		ClusterSize:    clusterSize,
		VirtualSize:    qcow2Header.VirtualSize,
		BackingFile:    backingFile,
		CryptMethod:    cryptoMethod,
		ImageEndOffset: imageEndOffset,
		Author:         consts.Author,
		Email:          consts.Email,
	}

	if Verbose {
		qcowOutputVerbose := structs.QCowOutputVerbose{
			QCowOutput:            qcowOutput,
			L1Size:                qcow2Header.L1Size,
			L1TableOffset:         qcow2Header.L1TableOffset,
			RefcountTableOffset:   qcow2Header.RefcountTableOffset,
			RefcountTableClusters: qcow2Header.RefcountTableClusters,
		}

		// todo check
		qcowOutputVerbose.DirtyBit = uint8(qcow2Header.IncompatibleFeatures & 0x8000000000000000)
		qcowOutputVerbose.CorruptBit = uint8(qcow2Header.IncompatibleFeatures & 0x4000000000000000)
		qcowOutputVerbose.ExternalDataFile = uint8(qcow2Header.IncompatibleFeatures & 0x2000000000000000)
		qcowOutputVerbose.CompressionType = uint8(qcow2Header.IncompatibleFeatures & 0x1000000000000000)
		qcowOutputVerbose.ExtendedL2 = uint8(qcow2Header.IncompatibleFeatures & 0x0800000000000000)

		qcowOutputVerbose.LazyRefCount = uint8(qcow2Header.CompatibleFeatures & 0x8000000000000000)

		qcowOutputVerbose.BitmapsExtension = uint8(qcow2Header.AutoClearFeatures & 0x8000000000000000)
		qcowOutputVerbose.RawExternalData = uint8(qcow2Header.AutoClearFeatures & 0x4000000000000000)

		qcowOutputVerbose.RefcountOrder = qcow2Header.RefcountOrder
		qcowOutputVerbose.HeaderLength = qcow2Header.HeaderLength
		qcowOutputVerbose.CompressionMethod = qcow2Header.CompressionType

		output, _ := json.MarshalIndent(qcowOutputVerbose, "", "    ")
		fmt.Printf("%s\n", output)
		return nil
	}

	output, _ := json.MarshalIndent(qcowOutput, "", "    ")
	fmt.Printf("%s\n", output)
	return nil
}
