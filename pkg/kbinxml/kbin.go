package kbinxml

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/orcaman/writerseeker"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const MagicNumber = 0xa042

type KbinEncoding struct {
	ID      int
	Name    string
	Decoder *encoding.Decoder
	Encoder *encoding.Encoder
}
type KbinNodeType struct {
	Name   string
	Size   int
	Count  int
	Signed bool
}
type ByteOffsets struct {
	offset4 int64
	offset2 int64
	offset1 int64
}

func (b *ByteOffsets) Align(reader io.Seeker) {
	offset, _ := reader.Seek(0, io.SeekCurrent)

	//fmt.Printf("OFFSET %0X %0X\n", offset, b)
	if offset%4 != 0 {
		offset = (offset - offset%4) + 4
	}
	newMaxOffset := int64(math.Max(float64(offset), float64(b.offset4)))
	//fmt.Println("NEW OFFSET", offset)
	if offset > b.offset4 {
		b.offset4 = newMaxOffset
	}
	//// LOG.Printf("ALIGN 2: %0X %v", b.offset2, b.offset2%4)
	if b.offset2%4 == 0 {
		b.offset2 = newMaxOffset
	}
	if b.offset1%4 == 0 {
		b.offset1 = newMaxOffset
	}
	//fmt.Println(b)
}

var nodeTypes = map[byte]KbinNodeType{
	TYPE_S8:     {"s8", 1, 1, true},
	TYPE_U8:     {"u8", 1, 1, false},
	TYPE_S16:    {"s16", 2, 1, true},
	TYPE_U16:    {"u16", 2, 1, false},
	TYPE_S32:    {"s32", 4, 1, true},
	TYPE_U32:    {"u32", 4, 1, false},
	TYPE_S64:    {"s64", 8, 1, true},
	TYPE_U64:    {"u64", 8, 1, false},
	TYPE_BIN:    {"bin", 1, 0, false},
	TYPE_STR:    {"str", 1, 0, false},
	TYPE_IP4:    {"ip4", 1, 4, false},
	TYPE_TIME:   {"time", 4, 1, false},
	TYPE_FLOAT:  {"float", 4, 1, true},
	TYPE_DOUBLE: {"double", 8, 1, false},
	TYPE_2S8:    {"2s8", 1, 2, true},
	TYPE_2U8:    {"2u8", 1, 2, false},
	TYPE_2S16:   {"2s16", 2, 2, false},
	TYPE_2U16:   {"2u16", 2, 2, false},
	TYPE_2S32:   {"2s32", 4, 2, true},
	TYPE_2U32:   {"2u32", 4, 2, false},
	TYPE_2S64:   {"2s64", 8, 2, true},
	TYPE_2U64:   {"2u64", 8, 2, false},
	TYPE_2F:     {"2f", 4, 2, false},
	TYPE_2D:     {"2d", 8, 2, false},
	TYPE_3S8:    {"3s8", 1, 3, true},
	TYPE_3U8:    {"3u8", 1, 3, false},
	TYPE_3S16:   {"3s16", 2, 3, false},
	TYPE_3U16:   {"3u16", 2, 3, false},
	TYPE_3S32:   {"3s32", 4, 3, true},
	TYPE_3U32:   {"3u32", 4, 3, false},
	TYPE_3S64:   {"3s64", 8, 3, true},
	TYPE_3U64:   {"3u64", 8, 3, false},
	TYPE_3F:     {"3f", 4, 3, false},
	TYPE_3D:     {"3d", 8, 3, false},
	TYPE_4S8:    {"4s8", 1, 4, true},
	TYPE_4U8:    {"4u8", 1, 4, false},
	TYPE_4S16:   {"4s16", 2, 4, false},
	TYPE_4U16:   {"4u16", 2, 4, false},
	TYPE_4S32:   {"4s32", 4, 4, true},
	TYPE_4U32:   {"4u32", 4, 4, false},
	TYPE_4S64:   {"4s64", 8, 4, true},
	TYPE_4U64:   {"4u64", 8, 4, false},
	TYPE_4F:     {"4f", 4, 4, false},
	TYPE_4D:     {"4d", 8, 4, false},
	TYPE_VS8:    {"vs8", 1, 16, true},
	TYPE_VU8:    {"vu8", 1, 16, false},
	TYPE_VS16:   {"vs16", 2, 8, true},
	TYPE_VU16:   {"vu16", 2, 8, false},
	TYPE_BOOL:   {"bool", 1, 1, false},
	TYPE_2B:     {"2b", 1, 2, false},
	TYPE_3B:     {"3b", 1, 3, false},
	TYPE_4B:     {"4b", 1, 4, false},
	TYPE_VB:     {"vb", 1, 16, false},
}

var numberTypes = map[string]byte{
	"b":    TYPE_BOOL,
	"bool": TYPE_BOOL,
	"s8":   TYPE_S8,
	"u8":   TYPE_U8,
	"s16":  TYPE_S16,
	"u16":  TYPE_U16,
	"s32":  TYPE_S32,
	"u32":  TYPE_U32,
	"s64":  TYPE_S64,
	"u64":  TYPE_U64,
	"time": TYPE_TIME,
}

const (
	TYPE_S8     = 2
	TYPE_U8     = 3
	TYPE_S16    = 4
	TYPE_U16    = 5
	TYPE_S32    = 6
	TYPE_U32    = 7
	TYPE_S64    = 8
	TYPE_U64    = 9
	TYPE_BIN    = 10
	TYPE_STR    = 11
	TYPE_IP4    = 12
	TYPE_TIME   = 13
	TYPE_FLOAT  = 14
	TYPE_DOUBLE = 15
	TYPE_2S8    = 16
	TYPE_2U8    = 17
	TYPE_2S16   = 18
	TYPE_2U16   = 19
	TYPE_2S32   = 20
	TYPE_2U32   = 21
	TYPE_2S64   = 22
	TYPE_2U64   = 23
	TYPE_2F     = 24
	TYPE_2D     = 25
	TYPE_3S8    = 26
	TYPE_3U8    = 27
	TYPE_3S16   = 28
	TYPE_3U16   = 29
	TYPE_3S32   = 30
	TYPE_3U32   = 31
	TYPE_3S64   = 32
	TYPE_3U64   = 33
	TYPE_3F     = 34
	TYPE_3D     = 35
	TYPE_4S8    = 36
	TYPE_4U8    = 37
	TYPE_4S16   = 38
	TYPE_4U16   = 39
	TYPE_4S32   = 40
	TYPE_4U32   = 41
	TYPE_4S64   = 42
	TYPE_4U64   = 43
	TYPE_4F     = 44
	TYPE_4D     = 45
	TYPE_VS8    = 48
	TYPE_VU8    = 49
	TYPE_VS16   = 50
	TYPE_VU16   = 51
	TYPE_BOOL   = 52
	TYPE_2B     = 53
	TYPE_3B     = 54
	TYPE_4B     = 55
	TYPE_VB     = 56
)
const (
	controlNodeStart = 1
	controlAttribute = 46
	controlNodeEnd   = 190
	controlFileEnd   = 191
)

const (
	EncodingNone = iota
	EncodingASCII
	EncodingISO_8859_1
	EncodingEUC_JP
	EncodingSHIFT_JIS
	EncodingUTF_8
)

var encodings = map[byte]*KbinEncoding{
	EncodingNone:       {0, "NONE", nil, nil},
	EncodingASCII:      {1, "ASCII", nil, nil},
	EncodingISO_8859_1: {2, "ISO-8859-1", charmap.ISO8859_1.NewDecoder(), charmap.ISO8859_1.NewEncoder()},
	EncodingEUC_JP:     {3, "EUC-JP", japanese.EUCJP.NewDecoder(), japanese.EUCJP.NewEncoder()},
	EncodingSHIFT_JIS:  {4, "SHIFT_JIS", japanese.ShiftJIS.NewDecoder(), japanese.ShiftJIS.NewEncoder()},
	EncodingUTF_8:      {5, "UTF-8", nil, nil},
}

// var LOG = LOG.New()

func DeserializeKbin(input []byte) ([]byte, *KbinEncoding, error) {
	// LOG.SetLevel(LOG.DebugLevel)
	reader := bytes.NewReader(input)

	//Check Magic Number
	magicNumber := make([]byte, 2)
	reader.Read(magicNumber)
	if binary.BigEndian.Uint16(magicNumber) != MagicNumber {
		fmt.Errorf("ERROR: Incorrect Magic Number: %v", magicNumber)
		return nil, nil, errors.New("incorrect magic number")
	}

	//Extract Encoding
	enc, _ := reader.ReadByte()
	documentEncoding := encodings[enc>>5]

	encChk, _ := reader.ReadByte()
	if 0xFF&^enc != encChk {
		fmt.Errorf("ERROR: Invalid encoding checksum, data corruption detected")
		return nil, nil, errors.New("invalid encoding checksum, data corruption detected")
	}

	nodes := getSegment(reader)
	data := getSegment(reader)

	document := etree.NewDocument()

	parseDocument(document, nodes, data, documentEncoding)

	parsedXML, err := document.WriteToBytes()
	return parsedXML, documentEncoding, err
}

func parseDocument(document *etree.Document, nodes []byte, data []byte, enc *KbinEncoding) {
	var curNode *etree.Element
	offsets := &ByteOffsets{0, 0, 0}
	nodeReader := bytes.NewReader(nodes)
	datareader := bytes.NewReader(data)

	for {
		nodeTypeId, _ := nodeReader.ReadByte()

		isArray := false
		if nodeTypeId&0x40 > 0 {
			isArray = true
			nodeTypeId -= 0x40
		}

		// LOG.Debugf("TypeID %v, Array: %v", nodeTypeId, isArray)
		nodeType := nodeTypes[nodeTypeId]
		// LOG.Debugf("OFFSETS 1:%0X 2:%0X 4:%0X", offsets.offset1, offsets.offset2, offsets.offset4)
		switch nodeTypeId {
		case controlNodeStart:
			name := parseSixbit(nodeReader)
			// LOG.Debugln("CONTROL_NODE_START")
			curNode = newNode(curNode, name)
		case controlAttribute:
			name := parseSixbit(nodeReader)
			// LOG.Debugln("CONTROL_ATTRIBUTE", name)
			datareader.Seek(offsets.offset4, io.SeekStart)
			attributeValue := getSegment(datareader)
			offsets.Align(datareader)
			value := transformString(attributeValue, enc)
			curNode.CreateAttr(name, value)
			// LOG.Debugf("NEW ATTR: [%s] %s", name, value)
		case controlNodeEnd:
			// LOG.Debugln("CONTROL_NODE_END")
			if curNode.Parent() != nil {
				curNode = curNode.Parent()
			}
		case controlFileEnd:
			// LOG.Debugln("CONTROL_FILE_END")
			document.AddChild(curNode)
			// LOG.Debugln(document.WriteToString())
			return
		case TYPE_STR:
			name := parseSixbit(nodeReader)
			// LOG.Debugf("TYPE_%s", nodeType.Name)
			datareader.Seek(offsets.offset4, io.SeekStart)
			valueBytes := getSegment(datareader)
			offsets.Align(datareader)
			value := transformString(valueBytes, enc)
			curNode = addLeaf(curNode, name, value, nodeType.Name, len(value), 0, nodeType)
		case TYPE_TIME, TYPE_FLOAT, TYPE_DOUBLE, TYPE_U8, TYPE_S8, TYPE_U16, TYPE_S16, TYPE_U32, TYPE_S32, TYPE_U64, TYPE_S64:
			name := parseSixbit(nodeReader)
			// LOG.Debugf("TYPE_%s %s", nodeType.Name, name)
			value, dataSize := extractNumber(datareader, offsets, nodeType, isArray)
			curNode = addLeaf(curNode, name, value, nodeType.Name, int(dataSize), int(dataSize)/nodeType.Size, nodeType)
		case TYPE_BOOL, TYPE_2B, TYPE_3B, TYPE_4B:
			name := parseSixbit(nodeReader)
			// LOG.Debugf("TYPE_%s", nodeType.Name)
			value, dataSize := extractBool(datareader, offsets, nodeType, isArray)
			curNode = addLeaf(curNode, name, value, nodeType.Name, int(dataSize), int(dataSize)/nodeType.Size, nodeType)
		case TYPE_BIN:
			name := parseSixbit(nodeReader)
			// LOG.Debugf("TYPE_%s", nodeType.Name)
			value, dataSize := extractBin(datareader, offsets)
			curNode = addLeaf(curNode, name, value, nodeType.Name, int(dataSize), 1, nodeType)
		case TYPE_IP4:
			name := parseSixbit(nodeReader)
			// LOG.Debugf("TYPE_%s", nodeType.Name)
			value := extractIp4(datareader, offsets)
			curNode = addLeaf(curNode, name, value, nodeType.Name, 0, 0, nodeType)
		default:
			// LOG.Errorln("CONTROL_UNKNOWN, ", nodeTypeId, nodeType.Name)
			return
		}
	}

}

func extractIp4(reader *bytes.Reader, offsets *ByteOffsets) (output string) {
	reader.Seek(offsets.offset4, io.SeekStart)
	ip := make([]byte, 4)
	reader.Read(ip)
	offsets.Align(reader)

	for _, v := range ip {
		output += fmt.Sprintf("%s.", strconv.Itoa(int(v)))
	}
	output = output[:len(output)-1]
	return
}

func extractBin(reader *bytes.Reader, offsets *ByteOffsets) (output string, dataSize uint32) {
	lengthBytes := make([]byte, 4)
	reader.Seek(offsets.offset4, io.SeekStart)
	reader.Read(lengthBytes)
	offsets.Align(reader)
	dataSize = binary.BigEndian.Uint32(lengthBytes)

	offset := offsets.getOffset(dataSize)

	data := make([]byte, dataSize)
	reader.Seek(*offset, io.SeekStart)
	reader.Read(data)
	offsets.Align(reader)
	return hex.EncodeToString(data), dataSize
}

func addLeaf(curNode *etree.Element, name string, value string, leafType string, size int, count int, nodeType KbinNodeType) *etree.Element {
	curNode = newNode(curNode, name)
	curNode.CreateAttr("__type", leafType)
	if size > nodeType.Size {
		curNode.CreateAttr("__size", strconv.Itoa(size))
	}
	if count > 1 {
		curNode.CreateAttr("__count", strconv.Itoa(count))
	}
	curNode.SetText(value)

	// LOG.Debugf("New Leaf: [%v] [%v] __type %v __size %v __count %v", name, value, leafType, size, count)
	return curNode
}

func extractBool(reader *bytes.Reader, offsets *ByteOffsets, nodeType KbinNodeType, array bool) (output string, dataSize uint32) {
	size := nodeType.Size
	count := nodeType.Count
	if array {
		lengthBytes := make([]byte, 4)
		reader.Seek(offsets.offset4, io.SeekStart)
		reader.Seek(0, io.SeekCurrent)
		reader.Read(lengthBytes)
		offsets.Align(reader)
		dataSize = binary.BigEndian.Uint32(lengthBytes)
		count = int(dataSize) / size
		// LOG.Debug("ARRAY /", size, dataSize, count, offset)
	} else {
		dataSize = uint32(nodeType.Size)
	}
	offset := offsets.getOffset(dataSize)
	reader.Seek(*offset, io.SeekStart)
	for i := 0; i < count; i++ {
		b, _ := reader.ReadByte()
		output += fmt.Sprintf("%b ", b)
	}
	*offset += int64(count * size)
	offsets.Align(reader)
	return output[:len(output)-1], dataSize
}

func (offsets *ByteOffsets) getOffset(dataSize uint32) *int64 {
	var offset *int64
	switch dataSize {
	case 1:
		offset = &offsets.offset1
	case 2:
		offset = &offsets.offset2
	default:
		offset = &offsets.offset4
	}
	return offset
}

func extractNumber(reader *bytes.Reader, offsets *ByteOffsets, nodeType KbinNodeType, array bool) (output string, dataSize uint32) {
	size := nodeType.Size
	count := nodeType.Count
	if array {
		lengthBytes := make([]byte, 4)
		reader.Seek(offsets.offset4, io.SeekStart)
		reader.Seek(0, io.SeekCurrent)
		reader.Read(lengthBytes)
		offsets.Align(reader)
		dataSize = binary.BigEndian.Uint32(lengthBytes)
		count = int(dataSize) / size
		// LOG.Debugf("Array found, data %v, length %v, count %v", lengthBytes, dataSize, count)
		// LOG.Debug("ARRAY /", size, dataSize, count, offset)
	} else {
		dataSize = uint32(nodeType.Size)
	}
	offset := offsets.getOffset(dataSize)
	reader.Seek(*offset, io.SeekStart)
	for i := 0; i < count; i++ {
		numberBytes := make([]byte, size)
		reader.Read(numberBytes)
		var number uint

		switch size {
		case 1:
			number = uint(numberBytes[0])
		case 2:
			number = uint(binary.BigEndian.Uint16(numberBytes))
		case 4:
			//fmt.Println(numberBytes)
			number = uint(binary.BigEndian.Uint32(numberBytes))
			//fmt.Println(number)
		case 8:
			number = uint(binary.BigEndian.Uint64(numberBytes))
		}

		if nodeType.Signed {
			numStr := ""
			switch size {
			case 1:
				numStr = fmt.Sprint(int8(number))
			case 2:
				numStr = fmt.Sprint(int16(number))
			case 4:
				numStr = fmt.Sprint(int32(number))
			case 8:
				numStr = fmt.Sprint(int64(number))
			}
			//if size == 4 {
			//	//fmt.Println(number, numStr, uint(binary.LittleEndian.Uint32(numberBytes)), hex.EncodeToString(numberBytes))
			//}
			if nodeType.Name == "float" || nodeType.Name == "double" {
				numStr = numStr[:len(numStr)-6] + "." + numStr[len(numStr)-6:]
			}
			output += fmt.Sprintf("%v ", numStr)
		} else {
			numStr := fmt.Sprint(number)
			if nodeType.Name == "float" || nodeType.Name == "double" {
				numStr = numStr[:len(numStr)-6] + "." + numStr[len(numStr)-6:]
			}
			output += fmt.Sprintf("%v ", numStr)
		}
	}
	*offset += int64(count * size)
	offsets.Align(reader)
	return output[:len(output)-1], dataSize
}
func transformString(input []byte, enc *KbinEncoding) string {
	if input[len(input)-1] == 0 {
		input = input[:len(input)-1]
	}
	if enc.Decoder == nil {
		return string(input)
	}
	transformer := transform.NewReader(bytes.NewReader(input), enc.Decoder)
	newString, _ := ioutil.ReadAll(transformer)
	return string(newString)
}

func newNode(curNode *etree.Element, name string) *etree.Element {
	// LOG.Debugf("NEW NODE: [%s]", name)
	newNode := etree.NewElement(name)
	if curNode != nil {
		curNode.AddChild(newNode)
	}
	return newNode
}

func parseSixbit(reader *bytes.Reader) string {
	length, _ := reader.ReadByte()
	if length == 0 {
		return ""
	}
	dataLength := int(math.Ceil(float64(length*6) / 8))
	data := make([]byte, dataLength)
	reader.Read(data)

	dataString := ""
	for _, v := range data {
		dataString += fmt.Sprintf("%08b", v)
	}

	output := ""
	for i := 0; i < int(length); i++ {
		chr, _ := strconv.ParseInt(dataString[0:6], 2, 64)
		dataString = dataString[6:]
		if chr < 10 {
			output += string(byte(chr + 48))
		} else if chr == 10 {
			output += ":"
		} else if chr < 37 {
			output += string(byte(chr + 54))
		} else if chr == 37 {
			output += "_"
		} else {
			output += string(byte(chr + 59))
		}
	}
	return output
}

func getSegment(reader *bytes.Reader) []byte {
	length := make([]byte, 4)
	reader.Read(length)
	segment := make([]byte, binary.BigEndian.Uint32(length))
	reader.Read(segment)
	return segment
}

func SerializeKbin(document *etree.Document, encoding int) (output []byte) {
	//// LOG.SetLevel(// LOG.DebugLevel)
	// Add magic numbers
	output = append(output, []byte{0xa0, 0x42}...)

	// Add encoding and checksum
	output = append(output, []byte{byte(encoding << 5), byte(0xFF &^ (encoding << 5))}...)
	iw := IterWalk(document)

	nodes := make([]byte, 0)
	data := make([]byte, 0)
	encoder := encodings[byte(encoding)]

	dataWriter := &writerseeker.WriterSeeker{}

	offsets := &ByteOffsets{}

	node, event := document.Root(), "start"
	for {
		//fmt.Println("event", event)
		//fmt.Printf("Node: Tag(%v) Text(%v) Attr(%v)\n", node.Tag, node.Text(), node.Attr)
		if event == "end" {
			nodes = append(nodes, controlNodeEnd|0x40)
			node, event = iw.Walk()
			//data, _ := ioutil.ReadAll(dataWriter.Reader())
			//fmt.Printf("%0X\n%0X\n\n", nodes, data)
			continue
		} else if event == "eof" {
			nodes = append(nodes, controlFileEnd|0x40)
			if len(nodes)%4 != 0 {
				nodes = append(nodes, bytes.Repeat([]byte{00}, 4-(len(nodes)%4))...)
			}
			data, _ = ioutil.ReadAll(dataWriter.Reader())
			data = append(data, bytes.Repeat([]byte{00}, int(offsets.offset4)-len(data))...)
			//fmt.Printf("%0X\n%0X\n\n", nodes, data)
			break
		}

		leafAttr := node.SelectAttr("__type")
		if leafAttr != nil { // leaf
			//fmt.Println("This is a leaf", node.Tag)
			// LOG.Debugf("%v %v OFFSETS 1:%0X 2:%0X 4:%0X", node.Tag, node.SelectAttr("__type").Value, offsets.offset1, offsets.offset2, offsets.offset4)
			switch node.SelectAttr("__type").Value {
			case "str":
				nodes = append(nodes, []byte{TYPE_STR, byte(len(node.Tag))}...)
				nodes = append(nodes, toSixbit(node.Tag)...)
				attrLength := make([]byte, 4)
				value, _ := encoder.Encoder.Bytes([]byte(node.Text()))
				value = append(value, 0x00)

				binary.BigEndian.PutUint32(attrLength, uint32(len(value)))
				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(attrLength)
				offsets.Align(dataWriter)

				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(value)
				offsets.Align(dataWriter)
			case "bin":
				nodes = append(nodes, []byte{TYPE_BIN, byte(len(node.Tag))}...)
				nodes = append(nodes, toSixbit(node.Tag)...)
				binLength := make([]byte, 4)
				data, _ := hex.DecodeString(node.Text())

				binary.BigEndian.PutUint32(binLength, uint32(len(data)))
				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(binLength)
				offsets.Align(dataWriter)

				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(data)
				offsets.Align(dataWriter)
			case "f", "float":
				nodes = append(nodes, []byte{TYPE_FLOAT, byte(len(node.Tag))}...)
				nodes = append(nodes, toSixbit(node.Tag)...)
				dataWriter.Seek(offsets.offset4, io.SeekStart)
				floatStr := node.Text()
				floatLst := strings.Split(floatStr, ".")
				floatMajor := floatLst[0]
				floatMinor := floatLst[1]
				if len(floatMinor) > 6 {
					floatMinor = floatMinor[0:6]
				} else if len(floatMinor) < 6 {
					floatMinor += strings.Repeat("0", 6-len(floatMinor))
				}
				floatInt, _ := strconv.Atoi(floatMajor + floatMinor)
				floatBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(floatBytes, uint32(floatInt))
				dataWriter.Write(floatBytes)
				offsets.Align(dataWriter)
			case "ip4":
				num := TYPE_IP4
				addresses := strings.Split(node.Text(), " ")
				realNum := byte(num)
				if len(addresses) > 1 {
					realNum |= 0x40
				}
				if len(addresses) > 1 {
					//fmt.Println("ARRAY")
					arrayLength := make([]byte, 4)
					binary.BigEndian.PutUint32(arrayLength, uint32(len(addresses)*4))
					dataWriter.Seek(offsets.offset4, io.SeekStart)
					dataWriter.Write(arrayLength)
					offsets.Align(dataWriter)
				}
				dataWriter.Seek(offsets.offset4, io.SeekStart)
				nodes = append(nodes, []byte{realNum, byte(len(node.Tag))}...)
				nodes = append(nodes, toSixbit(node.Tag)...)
				for _, addr := range addresses {
					addrLst := strings.Split(addr, ".")
					for _, a := range addrLst {
						aa, _ := strconv.ParseUint(a, 10, 8)
						dataWriter.Write([]byte{byte(aa)})
					}
				}
				offsets.Align(dataWriter)

			case "time", "b", "bool", "u8", "s8", "u16", "s16", "u32", "s32", "u64", "s64":
				num := numberTypes[node.SelectAttr("__type").Value]
				numbers := strings.Split(node.Text(), " ")
				// LOG.Debugln(node.Tag, numbers)
				realNum := num
				if len(numbers) > 1 {
					// LOG.Debugln(node.Tag, "IS_ARRAY")
					realNum |= 0x40
				}
				nodes = append(nodes, []byte{realNum, byte(len(node.Tag))}...)
				nodes = append(nodes, toSixbit(node.Tag)...)
				nt := nodeTypes[num]
				size := nt.Size * len(numbers)

				//fmt.Printf("%0X\n", offsets)
				if len(numbers) > 1 {
					//fmt.Println("ARRAY")
					// LOG.Debugln(node.Tag, "Size", size)
					arrayLength := make([]byte, 4)
					binary.BigEndian.PutUint32(arrayLength, uint32(size))
					dataWriter.Seek(offsets.offset4, io.SeekStart)
					// LOG.Debugln(node.Tag, "OFFSET, LENGTH", offsets.offset4, arrayLength)
					dataWriter.Write(arrayLength)
					offsets.Align(dataWriter)
				}
				// LOG.Debugln(node.Tag, "Sizer", size)
				offset := &offsets.offset1
				if size == 2 {
					offset = &offsets.offset2
				} else if size > 2 {
					offset = &offsets.offset4
				}
				//fmt.Println("SIZE", size, len(numbers), nt.Size)

				dataWriter.Seek(*offset, io.SeekStart)
				//fmt.Printf("%0X\n", offsets)
				for _, v := range numbers {
					var numberBytes = make([]byte, nt.Size)

					switch nt.Size {
					case 1:
						number, _ := strconv.Atoi(v)
						numberBytes[0] = byte(number)
					case 2:
						var number uint16
						if nt.Signed {
							num, _ := strconv.ParseInt(v, 10, 16)
							number = uint16(num)
						} else {
							num, _ := strconv.ParseUint(v, 10, 16)
							number = uint16(num)
						}
						binary.BigEndian.PutUint16(numberBytes, number)
					case 4:
						var number uint32
						if nt.Signed {
							num, _ := strconv.ParseInt(v, 10, 32)
							number = uint32(num)
						} else {
							num, _ := strconv.ParseUint(v, 10, 32)
							number = uint32(num)
						}
						binary.BigEndian.PutUint32(numberBytes, number)
					case 8:
						var number uint64
						if nt.Signed {
							num, _ := strconv.ParseInt(v, 10, 64)
							number = uint64(num)
						} else {
							number, _ = strconv.ParseUint(v, 10, 64)
						}

						//fmt.Printf("HERE %v %v %v\n", v, number, number)
						binary.BigEndian.PutUint64(numberBytes, number)
					}
					//					offs, _ := dataWriter.Seek(0, io.SeekCurrent)
					// LOG.Debugf("%v WRITING %0X", node.Tag, numberBytes)
					dataWriter.Write(numberBytes)
					if node.Tag == "s_coin" {
						// data, _ := ioutil.ReadAll(dataWriter.Reader())
						// LOG.Debugf("%0X", data)
					}
				}
				*offset += int64(size)
				//	offs, _ := dataWriter.Seek(0, io.SeekCurrent)
				//fmt.Printf("OFFS %0X\n", offs)
				offsets.Align(dataWriter)
				//fmt.Printf("OFFS %0X\n", offsets)
			default:
				panic(fmt.Sprintf("Unsupported type %v", node.SelectAttr("__type").Value))
			}
			for _, v := range node.Attr {
				if v.Key != "__type" && v.Key != "__count" && v.Key != "__size" {
					nodes = append(nodes, controlAttribute)
					nodes = append(nodes, byte(len(v.Key)))
					nodes = append(nodes, toSixbit(v.Key)...)

					attrLength := make([]byte, 4)
					value, _ := encoder.Encoder.Bytes([]byte(v.Value))
					value = append(value, 0x00)

					binary.BigEndian.PutUint32(attrLength, uint32(len(value)))
					dataWriter.Seek(offsets.offset4, io.SeekStart)
					dataWriter.Write(attrLength)
					offsets.Align(dataWriter)

					dataWriter.Seek(offsets.offset4, io.SeekStart)
					dataWriter.Write(value)
					offsets.Align(dataWriter)
				}
			}

		} else { // node
			nodes = append(nodes, []byte{controlNodeStart, byte(len(node.Tag))}...)
			nodes = append(nodes, toSixbit(node.Tag)...)
			for _, v := range node.Attr {
				nodes = append(nodes, controlAttribute)
				nodes = append(nodes, byte(len(v.Key)))
				nodes = append(nodes, toSixbit(v.Key)...)

				attrLength := make([]byte, 4)
				value, _ := encoder.Encoder.Bytes([]byte(v.Value))
				value = append(value, 0x00)

				binary.BigEndian.PutUint32(attrLength, uint32(len(value)))
				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(attrLength)
				offsets.Align(dataWriter)

				dataWriter.Seek(offsets.offset4, io.SeekStart)
				dataWriter.Write(value)
				offsets.Align(dataWriter)
			}
		}

		//data, _ := ioutil.ReadAll(dataWriter.Reader())
		//fmt.Printf("%0X\n%0X\n\n", nodes, data)

		node, event = iw.Walk()
	}
	nodeLength := make([]byte, 4)
	dataLength := make([]byte, 4)
	binary.BigEndian.PutUint32(nodeLength, uint32(len(nodes)))
	binary.BigEndian.PutUint32(dataLength, uint32(len(data)))
	output = append(output, nodeLength...)
	output = append(output, nodes...)
	output = append(output, dataLength...)
	output = append(output, data...)
	return
}

func toSixbit(text string) (output []byte) {
	bitstring := ""
	for _, v := range text {
		var char int32
		if match, _ := regexp.Match("[0-9]", []byte{byte(v)}); match {
			char = v - 48
		} else if match, _ := regexp.Match(":", []byte{byte(v)}); match {
			char = 10
		} else if match, _ := regexp.Match("[A-Z]", []byte{byte(v)}); match {
			char = v - 54
		} else if match, _ := regexp.Match("_", []byte{byte(v)}); match {
			char = 37
		} else if match, _ := regexp.Match("[a-z]", []byte{byte(v)}); match {
			char = v - 59
		} else {
			panic(fmt.Sprintf("invalid sixbit char %v", v))
		}
		bitstring += fmt.Sprintf("%08b", char)[2:]
	}
	if len(bitstring)%8 != 0 {
		extra := 8 - (len(bitstring) % 8)
		bitstring += strings.Repeat("0", extra)
	}
	for i := 0; i < len(bitstring); i += 8 {
		outbyte, _ := strconv.ParseUint(bitstring[i:i+8], 2, 64)
		output = append(output, byte(outbyte))
	}
	return
}
