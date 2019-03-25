package rpc

import (
	"encoding/binary"
	"encoding/json"
	"math"

	"github.com/golang/protobuf/proto"
)

func convertToBytes(order binary.ByteOrder, valList ...interface{}) ([]byte, error) {
	data := make([]byte, 0, 1024)
	tmpVal := make([]byte, 8)
	for _, item := range valList {
		switch item.(type) {
		case int:
			{
				order.PutUint32(tmpVal, uint32(item.(int)))
				data = append(data, tmpVal[:4]...)
			}
		case int8:
			{
				data = append(data, byte(item.(int8)))
			}
		case int16:
			{
				order.PutUint16(tmpVal, uint16(item.(int16)))
				data = append(data, tmpVal[:2]...)
			}
		case int32:
			{
				order.PutUint32(tmpVal, uint32(item.(int)))
				data = append(data, tmpVal[:4]...)
			}
		case int64:
			{
				order.PutUint64(tmpVal, uint64(item.(int64)))
				data = append(data, tmpVal[:8]...)
			}
		case uint:
			{
				order.PutUint32(tmpVal, uint32(item.(uint)))
				data = append(data, tmpVal[:4]...)
			}
		case uint8:
			{
				data = append(data, byte(item.(uint8)))
			}
		case uint16:
			{
				order.PutUint16(tmpVal, item.(uint16))
				data = append(data, tmpVal[:2]...)
			}
		case uint32:
			{
				order.PutUint32(tmpVal, item.(uint32))
				data = append(data, tmpVal[:4]...)
			}
		case uint64:
			{
				order.PutUint64(tmpVal, item.(uint64))
				data = append(data, tmpVal[:8]...)
			}
		case float32:
			{
				var val = math.Float32bits(item.(float32))
				order.PutUint32(tmpVal, val)
				data = append(data, tmpVal[:4]...)
			}
		case float64:
			{
				var val = math.Float64bits(item.(float64))
				order.PutUint64(tmpVal, val)
				data = append(data, tmpVal[:8]...)
			}
		case string:
			{
				var val = []byte(item.(string))
				order.PutUint32(tmpVal, uint32(len(val)))
				data = append(data, tmpVal[:4]...)
				data = append(data, val...)
			}
			//****************** 指针处理

		case []int:
			{
				tmpData := item.([]int)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint32(tmpVal, uint32(tmpData[i]))
					data = append(data, tmpVal[:4]...)
				}
			}
		case []int8:
			{
				tmpData := item.([]int8)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					data = append(data, byte(tmpData[i]))
				}
			}
		case []int16:
			{
				tmpData := item.([]int16)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint16(tmpVal, uint16(tmpData[i]))
					data = append(data, tmpVal[:2]...)
				}
			}
		case []int32:
			{
				tmpData := item.([]int32)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint32(tmpVal, uint32(tmpData[i]))
					data = append(data, tmpVal[:4]...)
				}
			}
		case []int64:
			{
				tmpData := item.([]int64)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint64(tmpVal, uint64(tmpData[i]))
					data = append(data, tmpVal[:8]...)
				}
			}
		case []uint:
			{
				tmpData := item.([]uint)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint32(tmpVal, uint32(tmpData[i]))
					data = append(data, tmpVal[:4]...)
				}
			}
		case []uint8:
			{
				tmpData := item.([]uint8)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					data = append(data, byte(tmpData[i]))
				}
			}
		case []uint16:
			{
				tmpData := item.([]uint16)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint16(tmpVal, tmpData[i])
					data = append(data, tmpVal[:2]...)
				}
			}
		case []uint32:
			{
				tmpData := item.([]uint32)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint32(tmpVal, tmpData[i])
					data = append(data, tmpVal[:4]...)
				}
			}
		case []uint64:
			{
				tmpData := item.([]uint64)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					order.PutUint64(tmpVal, tmpData[i])
					data = append(data, tmpVal[:8]...)
				}
			}
		case []float32:
			{
				tmpData := item.([]float32)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					var val = math.Float32bits(tmpData[i])
					order.PutUint32(tmpVal, val)
					data = append(data, tmpVal[:4]...)
				}
			}
		case []float64:
			{
				tmpData := item.([]float64)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					var val = math.Float64bits(tmpData[i])
					order.PutUint64(tmpVal, val)
					data = append(data, tmpVal[:8]...)
				}
			}
		case []string:
			{
				tmpData := item.([]string)
				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				for i := 0; i < len(tmpData); i++ {
					var val = []byte(tmpData[i])
					order.PutUint32(tmpVal, uint32(len(val)))
					data = append(data, tmpVal[:4]...)
					data = append(data, val...)
				}
			}
		case proto.Message:
			{
				tmpData, err := proto.Marshal(item.(proto.Message))
				if err != nil {
					return nil, err
				}

				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				data = append(data, tmpData...)
			}
		case *proto.Message:
			{
				tmpData, err := proto.Marshal(*item.(*proto.Message))
				if err != nil {
					return nil, err
				}

				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				data = append(data, tmpData...)
			}
		case []proto.Message:
			{
				tmpData, err := proto.Marshal(item.(proto.Message))
				if err != nil {
					return nil, err
				}

				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				data = append(data, tmpData...)
			}
		case error:
			{
				var val = []byte(item.(error).Error())
				order.PutUint32(tmpVal, uint32(len(val)))
				data = append(data, tmpVal[:4]...)
				data = append(data, val...)
			}
		default:
			{
				tmpData, err := json.Marshal(item)
				if err != nil {
					return nil, err
				}

				order.PutUint32(tmpVal, uint32(len(tmpData)))
				data = append(data, tmpVal[:4]...)

				data = append(data, tmpData...)
			}
		}
	}

	return data, nil
}
