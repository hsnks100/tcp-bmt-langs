#pragma once
#include <inttypes.h>
#include <memory.h>
#pragma pack(push, 1)
typedef struct PKT_HDRtag
{
    int32_t     ServiceCode;           // 서비스코드
    int32_t     ResultCode;            // 리턴 코드
    int32_t     DataLen;               // 데이터 길이
    uint32_t    Checksum;              // 헤더 체크섬, iServiceCode ^ iResultCode ^ iDataLen
	uint8_t body[0];
} PKT_HDR;
typedef struct Body_ {
    char                     szKey[32];			//	bno_resolution_bitrate
    uint64_t frameNumber;
    uint8_t frameType; // 'A' or 'V'
    uint8_t                     Frame[0];			//	프레임 데이터
} Body;
#pragma pack(pop)
