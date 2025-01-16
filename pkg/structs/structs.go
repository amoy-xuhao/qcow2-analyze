package structs

type QCowHeader struct {
	/* QCOW magic string ("QFI\xfb") */
	Magic string `json:"magic"`
	/* Version number (valid values are 2 and 3) */
	Version uint32 `json:"version"`
	/* Offset into the image file at which the backing file name is stored */
	BackingFileOffset uint64 `json:"backing_file_offset"`
	/* Length of the backing file name in bytes (less than 1024) */
	BackingFileSize uint32 `json:"backing_file_size"`
	/* 1 << cluster_bits is the cluster size */
	ClusterBits uint32 `json:"cluster_bits"`
	/* virtual size in bytes */
	VirtualSize uint64 `json:"virtual_size"`
	/* 0 for no encryption; 1 for AES encryption; 2 for LUKS encryption */
	CryptMethod uint32 `json:"crypt_method"`
	/* Number of entries(L2 table) in the active L1 table */
	L1Size uint32 `json:"l1_size"`
	/* Offset into the image file at which the active L1 table starts */
	L1TableOffset uint64 `json:"l1_table_offset"`
	/* Offset into the image file at which the refcount table starts */
	RefcountTableOffset uint64 `json:"refcount_table_offset"`
	/* Number of clusters that the refcount table occupies */
	RefcountTableClusters uint32 `json:"refcount_table_clusters"`
	/* Number of snapshots(internal snapshot) contained in the image */
	NumberSnapshots uint32 `json:"number_snapshots"`
	/* Offset into the image file at which the snapshot table starts */
	SnapshotsOffset uint64 `json:"snapshots_offset"`

	/* The following fields are only valid for version >= 3 */
	IncompatibleFeatures uint64 `json:"incompatible_features"`
	CompatibleFeatures   uint64 `json:"compatible_features"`
	AutoClearFeatures    uint64 `json:"autoclear_features"`
	/* refcount_bits = 1 << refcount_order */
	RefcountOrder uint32 `json:"refcount_order"`
	HeaderLength  uint32 `json:"header_length"`

	/* Additional fields */
	CompressionType uint8 `json:"compression_type"`
	/* header must be a multiple of 8 */
	Padding [7]uint8 `json:"padding,omitempty"`
}

type QCowOutput struct {
	QCowVersion uint32 `json:"qcow_version"`
	BackingFile string `json:"backing_file"`
	ClusterSize uint32 `json:"cluster_size"`
	VirtualSize uint64 `json:"virtual_size"`
	CryptMethod string `json:"crypt_method"`

	/* internal snapshots */
	Snapshots []string `json:"snapshots"`

	/* calculate by refcount table */
	ImageEndOffset uint64 `json:"image_end_offset"`

	// author of the tools
	Author string `json:"author"`
	Email  string `json:"email"`
}

type QCowOutputVerbose struct {
	QCowOutput
	L1Size                uint32 `json:"l1_size"`
	L1TableOffset         uint64 `json:"l1_table_offset"`
	RefcountTableOffset   uint64 `json:"refcount_table_offset"`
	RefcountTableClusters uint32 `json:"refcount_table_clusters"`

	/* version 3 features */
	/* incompatible features */
	DirtyBit         uint8 `json:"dirty_bit"`
	CorruptBit       uint8 `json:"corrupt_bit"`
	ExternalDataFile uint8 `json:"external_data_file"`
	CompressionType  uint8 `json:"compression_type"`
	ExtendedL2       uint8 `json:"extended_l2"`

	/* compatible features */
	LazyRefCount uint8 `json:"lazy_refcount"`

	/* auto clear features */
	BitmapsExtension uint8 `json:"bitmaps_extension"`
	RawExternalData  uint8 `json:"raw_external_data"`

	RefcountOrder     uint32 `json:"refcount_order"`
	HeaderLength      uint32 `json:"header_length"`
	CompressionMethod uint8  `json:"compression_method"`
}
