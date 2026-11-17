/////////////////////////////////////////////////////////////////////////////
//
//  @file efml.h
//
//  @brief API interface of Enflame Management Library
//
//  Enflame Tech, All Rights Reserved. 2019-2026 Copyright (C)
//
/////////////////////////////////////////////////////////////////////////////

#if !defined(_EFML_H_)
#define _EFML_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define MAX_CHAR_BUFF_LEN (128)

typedef enum {
    EFML_SUCCESS                        = 0,  // Success
    EFML_ERROR_UNINITIALIZED            = 1,  // Error since uninitialized.
    EFML_ERROR_INVALID_ARGUMENT         = 2,
    EFML_ERROR_NOT_SUPPORTED            = 3,
    EFML_ERROR_LIBRARY_NOT_FOUND        = 4,
    EFML_ERROR_INVALID_ERROR_CODE       = 5,
    EFML_ERROR_DRIVER_NOT_LOADED        = 6,
    EFML_ERROR_GCU_LARE_PORT_NUMBER_ERR = 7,
    EFML_ERROR_INVALID_INPUT            = 8,
    EFML_ERROR_FUNCTION_NOT_FOUND       = 9,
    EFML_ERROR_OPEN_DRIVER_VERSION      = 10,
    EFML_ERROR_DRIVER_NOT_COMPATIBLE    = 11,
    EFML_ERROR_NO_DEVICE                = 12,
    EFML_ERROR_DEPRECATION              = 13,
    EFML_ERROR_TIMEOUT                  = 253,
    EFML_ERROR_FAIL                     = 254,
    EFML_ERROR_MAX                      = 255,
} efmlReturn_t;

typedef enum {
    EFML_LINK_SPEED_GEN1 = 1,
    EFML_LINK_SPEED_GEN2 = 2,
    EFML_LINK_SPEED_GEN3 = 3,
    EFML_LINK_SPEED_GEN4 = 4,
    EFML_LINK_SPEED_GEN5 = 5,
} efmlPcieSpeed_t;

typedef enum {
    EFML_GCU_LARE_LINK_SPEED_GEN1      = 1,
    EFML_GCU_LARE_LINK_SPEED_GEN2      = 2,
    EFML_GCU_LARE_LINK_SPEED_GEN3      = 3,
    EFML_GCU_LARE_LINK_SPEED_GEN4      = 4,
    EFML_GCU_LARE_LINK_SPEED_GEN5      = 5,
    EFML_GCU_LARE_LINK_SPEED_ESM_2P5GT = 0x10,
    EFML_GCU_LARE_LINK_SPEED_ESM_5GT,
    EFML_GCU_LARE_LINK_SPEED_ESM_8GT,
    EFML_GCU_LARE_LINK_SPEED_ESM_16GT,
    EFML_GCU_LARE_LINK_SPEED_ESM_20GT,
    EFML_GCU_LARE_LINK_SPEED_ESM_25GT
} efmlGcuLareSpeed_t;


typedef enum {
    EFML_GCU_LARE_LINK_WIDTH_X1  = 1,
    EFML_GCU_LARE_LINK_WIDTH_X2  = 2,
    EFML_GCU_LARE_LINK_WIDTH_X4  = 4,
    EFML_GCU_LARE_LINK_WIDTH_X8  = 8,
    EFML_GCU_LARE_LINK_WIDTH_X16 = 16,
} efmlGcuLareWidth_t;

typedef enum {
    EFML_GCU_LARE_PORT_RC = 0,
    EFML_GCU_LARE_PORT_EP = 1,
} efmlGcuLarePortType_t;

typedef enum {
    EFML_EVENT_UNKNOWN          = 0,
    EFML_EVENT_DTU_SUSPEND      = 3,
    EFML_EVENT_DTU_RESET_START  = 10,
    EFML_EVENT_DTU_RESET_FINISH = 11,
    EFML_EVENT_DTU_EID          = 20,
} efmlEventType_t;

typedef struct {
    uint32_t        event_id;
    efmlEventType_t event_type;
    char            event_msg[MAX_CHAR_BUFF_LEN];
} efmlEvent_t;


typedef enum {
    EFML_LINK_WIDTH_X1  = 1,
    EFML_LINK_WIDTH_X2  = 2,
    EFML_LINK_WIDTH_X4  = 4,
    EFML_LINK_WIDTH_X8  = 8,
    EFML_LINK_WIDTH_X16 = 16,
} efmlPcieWidth_t;

typedef enum {
    GCU200 = 0,
    GCU210 = 1,
    GCU300 = 2,
    GCU400 = 3,
    GCU410 = 4,
    GCU450 = 5,
} efmlHwArch_t;

typedef enum {
    UNKNOWN  = 0,
    GCU_ARCH_200  = 20, /* pavo */
    GCU_ARCH_210  = 21, /* dorado */
    GCU_ARCH_300  = 30, /* scorpio */
    GCU_ARCH_400  = 40, /* libra */
    GCU_ARCH_410  = 41, /* libra B0 */
    GCU_ARCH_450  = 45, /* libra H */
} efmlHwArchV2_t;

typedef enum {
    CABLE_NOT_QUALIFIED = 0,
    CABLE_QUALIFIED     = 1,
} efmlCableQualifySts_t;

typedef enum {
    EFML_HBM_SCAN_INVALID = 0,
    EFML_HBM_SCAN_START   = 1,
} efmlHbmScanType_t;

typedef enum {
    EFML_ACTIVE  = 1,
    EFML_SUSPEND = 5,
    EFML_SLEEP   = 7,
    EFML_ULP     = 9,
} efmlPowerMode_t;

typedef enum {
    EFML_GCU_VIRT_MODE_HOST,
    EFML_GCU_VIRT_MODE_VGCU,
    EFML_GCU_VIRT_MODE_MDEV,
    EFML_GCU_VIRT_MODE_SRIOV_PF,
    EFML_GCU_VIRT_MODE_SRIOV_VF,
} efmlGcuVirtMode_t;

/**
* @brief Enflame data structure for device thermal info.
*
*/
typedef struct {
   float cur_dev_temp;
   float cur_hbm0_temp;
   float cur_hbm1_temp;
}efmlDevThermalInfo_t;

#define MAX_DEVICE_COUNT (128)
#define MAX_PROCESS_COUNT (64)

/**
 * @brief Enflame data structure for device info.
 *
 */
 typedef struct {
    char     name[MAX_CHAR_BUFF_LEN];
    uint32_t vendor_id;
    uint32_t device_id;
    uint32_t domain_id;
    uint32_t bus_id;
    uint32_t dev_id;
    uint32_t func_id;
    uint32_t logic_id;
} efmlDeviceInfo_t;

/**
* @brief Enflame data structure for device power info.
*
*/
typedef struct {
    float pwr_capability;
    float cur_pwr_consumption;
} efmlDevPowerInfo_t;

/**
* @brief Enflame data structure for device memory info.
*
*/
typedef struct {
    uint64_t mem_total_size;  // in bytes
    uint32_t mem_used;        // in MiB
} efmlDevMemInfo_t;

/**
*@brief Enflame data structure for HBM usage info.
*
*/
typedef struct {
    uint64_t hbm_total;
    uint64_t hbm_free;
    uint64_t hbm_used;
} efmlHbmUsage_t;

/**
*@brief Enflame data structure for CQM usage info.
*
*/
typedef struct {
    uint32_t cqm_total;
    uint32_t cqm_init;
    uint32_t cqm_idle;
    uint32_t cqm_busy;
    uint32_t cqm_masked;
    uint32_t cqm_hwerr;
} efmlCqmUsage_t;

/**
*@brief Enflame data structure for SIP usage info.
*
*/
typedef struct {
    uint32_t sip_total;
    uint32_t sip_init;
    uint32_t sip_idle;
    uint32_t sip_busy;
    uint32_t sip_masked;
    uint32_t sip_hwerr;
} efmlSipUsage_t;

 /**
* @brief Enflame data structure for DTU usage info.
*
*/
typedef struct {
        efmlHbmUsage_t hbm;
        efmlSipUsage_t sip;
        efmlCqmUsage_t cqm;
}efmlDtuUsage_t;

/**
* @brief Enflame data structure for cluster memory info.
*
*/
typedef struct {
    uint64_t mem_total_size;
    uint32_t mem_used;
} efmlClusterHbmMemInfo_t;

/**
* @brief Enflame data structure for device volt info.
*
*/
typedef struct {
    float vdd_dtu;
    float vdd_soc;
    float vdd_hbmqc;
} efmlDevVoltInfo_t;

/**
 * @brief Enflame data structure for Pcie link info.
 *
 */
 typedef struct {
    efmlPcieSpeed_t link_speed;
    efmlPcieSpeed_t max_link_speed;
    efmlPcieWidth_t link_width;
    efmlPcieWidth_t max_link_width;
} efmlPcieLinkInfo_t;

/**
* @brief Enflame data structure for GCU-LARE link info.
*
*/
typedef struct {
    efmlGcuLareSpeed_t link_speed;
    efmlGcuLareSpeed_t max_link_speed;
    efmlGcuLareWidth_t link_width;
    efmlGcuLareWidth_t max_link_width;
} efmlGcuLareLinkInfo_t;

/**
*@brief Enflame data structure for GCU-LARE Throughput info.
*
*/
typedef struct {
    float    tx_throughput;
    float    rx_throughput;
    uint64_t tx_nak;
    uint64_t rx_nak;
} efmlGcuLareThroughputInfo_t;

/**
 * @brief Enflame data structure for GCU-LARE port info.
 *
 */
 typedef struct {
    uint32_t          connected;
    char              uuid[16];
    uint32_t          vendor_id;
    uint32_t          device_id;
    uint32_t          domain_id;
    uint32_t          bus_id;
    uint32_t          dev_id;
    uint32_t          func_id;
    uint32_t          port_id;
    efmlGcuLarePortType_t port_type;

    uint32_t          remote_card_id;
    char              remote_uuid[16];
    uint32_t          remote_vendor_id;
    uint32_t          remote_device_id;
    uint32_t          remote_domain_id;
    uint32_t          remote_bus_id;
    uint32_t          remote_dev_id;
    uint32_t          remote_func_id;
    uint32_t          remote_port_id;
    efmlGcuLarePortType_t remote_port_type;
} efmlGcuLarePortInfo_t;

/**
*@brief Enflame data structure for Rma details.
*
*/
typedef struct {
    bool     is_dtu_support_rma;
    bool     flags;  // true: hint, false: not hint
    uint32_t dbe_count;
} efmlRmaDetails_t;


/**
*@brief Enflame data structure for Rma details.
*
*/
typedef struct {
    bool     is_dtu_support_rma;
    bool     flags;  // true: hint, false: not hint
    uint32_t total_dbe_count;
    uint32_t mc_count;
    uint32_t per_mc_dbe_count[32];  // only 12 is used now
} efmlRmaDetailsV2_t;

/**
*@brief Enflame data structure for Decoder info.
*
*/
typedef struct {
    uint32_t decoder_inst_num;    // decoder instance number
    float    decoder_resolution;  // decoder resolution
    float    decoder_fps;         // decoder fps
} efmlDecoderCap_t;

typedef struct {
    uint32_t pid;           /* process pid         */
    uint64_t dev_mem_usage; /* device memory usage */
    uint64_t sys_mem_usage; /* system memory usage */
} efmlProcessInfo_t;

typedef enum {
    OP_SPACE_GCU,
    OP_SPACE_PF,
    OP_SPACE_CTL,
} efmlOperateSpace_t;

typedef struct {
    uint32_t encoder_usage;
    uint32_t decoder_usage;
} efmlVpuUsage_t;

/**
*@brief Enflame data structure for Rma status.
*
*/
typedef struct {
    bool is_dtu_support_rma;
    bool flags;  // true: hint, false: not hint
} efmlRmaStatus_t;

typedef struct {
    bool     enabled;  // true: enabled               false: disabled
    bool     pending;  // true: pending               false: no pending
    bool     pdblack;  // true: pending blacklist yes false: no
    uint32_t ecnt_sb;  // single bit error count
    uint32_t ecnt_db;  // double bit error count
} efmlEccStatus_t;

/**
* @brief Enflame data structure for Pcie Throughput info.
*
*/
typedef struct {
    float    tx_throughput;
    float    rx_throughput;
    uint64_t tx_nak;
    uint64_t rx_nak;
} efmlPcieThroughputInfo_t;

typedef struct {
    float cur_asic_temp;
    float cur_mem_temp;
    float cur_board_temp;
} efmlDevThermalInfoV2_t;

/**
* @brief Enflame data structure for device clock info.
*
*/
typedef struct {
    uint32_t cur_hbm_clock;
    uint32_t cur_dtu_clock;
} efmlDevClkInfo_t;


typedef enum {
    EFML_ESL_LINK_SPEED_GEN1      = 1,
    EFML_ESL_LINK_SPEED_GEN2      = 2,
    EFML_ESL_LINK_SPEED_GEN3      = 3,
    EFML_ESL_LINK_SPEED_GEN4      = 4,
    EFML_ESL_LINK_SPEED_GEN5      = 5,
    EFML_ESL_LINK_SPEED_ESM_2P5GT = 0x10,
    EFML_ESL_LINK_SPEED_ESM_5GT,
    EFML_ESL_LINK_SPEED_ESM_8GT,
    EFML_ESL_LINK_SPEED_ESM_16GT,
    EFML_ESL_LINK_SPEED_ESM_20GT,
    EFML_ESL_LINK_SPEED_ESM_25GT
} efmlEslSpeed_t;

typedef enum {
    EFML_ESL_LINK_WIDTH_X1  = 1,
    EFML_ESL_LINK_WIDTH_X2  = 2,
    EFML_ESL_LINK_WIDTH_X4  = 4,
    EFML_ESL_LINK_WIDTH_X8  = 8,
    EFML_ESL_LINK_WIDTH_X16 = 16,
} efmlEslWidth_t;

typedef enum {
    EFML_ESL_PORT_RC = 0,
    EFML_ESL_PORT_EP = 1,
} efmlEslPortType_t;

/**
* @brief Enflame data structure for esl link info.
*
*/
typedef struct {
    efmlEslSpeed_t link_speed;
    efmlEslSpeed_t max_link_speed;
    efmlEslWidth_t link_width;
    efmlEslWidth_t max_link_width;
} efmlEslLinkInfo_t;

/**
*@brief Enflame data structure for Esl Throughput info.
*
*/
typedef struct {
    float    tx_throughput;
    float    rx_throughput;
    uint64_t tx_nak;
    uint64_t rx_nak;
} efmlEslThroughputInfo_t;


/**
 * @brief Enflame data structure for esl port info.
 *
 */
typedef struct {
    uint32_t          connected;
    char              uuid[16];
    uint32_t          vendor_id;
    uint32_t          device_id;
    uint32_t          domain_id;
    uint32_t          bus_id;
    uint32_t          dev_id;
    uint32_t          func_id;
    uint32_t          port_id;
    efmlEslPortType_t port_type;

    uint32_t          remote_card_id;
    char              remote_uuid[16];
    uint32_t          remote_vendor_id;
    uint32_t          remote_device_id;
    uint32_t          remote_domain_id;
    uint32_t          remote_bus_id;
    uint32_t          remote_dev_id;
    uint32_t          remote_func_id;
    uint32_t          remote_port_id;
    efmlEslPortType_t remote_port_type;
} efmlEslPortInfo_t;

/**
 * @brief Enflame Management Library initialization
 *
 * @return efmlReturn_t
 */
EFML_EXPORT efmlReturn_t EfmlInit(bool no_driver);

 /**
  * @brief Enflame Management Library Shutdown.
  *
  */
EFML_EXPORT void EfmlShutdown();

 /**
  * @brief Enflame Management Library map error code to string.
  *
  */
EFML_EXPORT efmlReturn_t EfmlErrorString(efmlReturn_t result, char *p_error_str);

 /**
  * @brief Enflame Management Library get driver version info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDriverVer(char *p_driver_ver);

 /**
  * @brief Enflame Management Library get efml lib version info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetLibVer(char *p_self_ver);

 /**
  * @brief Enflame Management Library get enflame driver status and config access path.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDriverAccessPoint(char *p_enflame_driver_ap);

 /**
  * @brief Get the total number of supported devices
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevCount(uint32_t *dev_count);

 /**
  * @brief Get the total number of enabled virtual devices of a physical device
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetVdevCount(uint32_t dev_idx, uint32_t *vdev_count);

 /**
  * @brief Get the virtual devices index in operate system
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetVdevList(uint32_t dev_idx, uint32_t *vdev_ids, uint32_t *count);

 /**
  * @brief Get the total number of maximum supported virtual devices of a physical device
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetMaxVdevCount(uint32_t dev_idx, uint32_t *vdev_count);

 /**
  * @brief Enflame Management Library get the device name.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevName(uint32_t dev_idx, char *p_name);

 /**
  * @brief Enflame Management Library get the device temperature.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevTempV2(uint32_t dev_idx, efmlDevThermalInfoV2_t *p_temp);

 /**
  * @brief Enflame Management Library get the device voltage.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevVolt(uint32_t dev_idx, efmlDevVoltInfo_t *p_volt);

 /**
  * @brief Enflame Management Library get the device current power consumption.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevPwr(uint32_t dev_idx, efmlDevPowerInfo_t *p_pwr);

 /**
  * @brief Enflame Management Library get the device mem info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevMem(uint32_t dev_idx, efmlDevMemInfo_t *p_mem);

 /**
  * @brief Enflame Management Library get the device usage.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevDtuUsage(uint32_t dev_idx, float *p_data);

 /**
  * @brief Enflame Management Library get the virtual device mem info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetVdevMem(uint32_t dev_idx, uint32_t vdev_idx, efmlDevMemInfo_t *p_mem);

 /**
  * @brief Enflame Management Library get the virtual device usage.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetVdevDtuUsage(uint32_t dev_idx, uint32_t vdev_idx, float *p_data);

EFML_EXPORT efmlReturn_t EfmlGetDevSipUsage(uint32_t dev_idx, float *p_data);

 /**
  * @brief Enflame Management Library get the device usage from background sample thread.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevDtuUsageAsync(uint32_t dev_idx, float *p_data);

 /**
  * @brief Enflame Management Library get if dtu is in low power mode.
  *
  */
 EFML_DEPRECATED_EXPORT efmlReturn_t EfmlGetDevIsLowPowerMode(uint32_t dev_idx, bool *is_low_power_mode);

 /**
  * @brief Enflame Management Library get if dtu is health.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevIsHealth(uint32_t dev_idx, bool *is_health);

 /**
  * @brief Enflame Management Library get the device max clock frequency.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetMaxFreq(uint32_t dev_idx, uint32_t *max_freq_mhz);

 /**
  * @brief Enflame Management Library set the device max clock frequency.
  *
  */
EFML_EXPORT efmlReturn_t EfmlSetMaxFreq(uint32_t dev_idx, uint32_t max_freq_mhz);

 /**
  * @brief Enflame Management Library get the device clock info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevClk(uint32_t dev_idx, efmlDevClkInfo_t *p_clk);

 /**
  * @brief Enflame Management Library get the memory clock info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetMemClk(uint32_t dev_idx, uint32_t *p_clk_mhz);

 /**
  * @brief Enflame Management Library get device info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevInfo(uint32_t dev_idx, efmlDeviceInfo_t *p_info);

 /**
  * @brief Enflame Management Library display device topology.
  *
  */
EFML_EXPORT efmlReturn_t EfmlDisplayDevTop(uint32_t dev_idx);

 /**
  * @brief Enflame Management Library get firmware version info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetFwVersion(uint32_t dev_idx, char *fw_ver);

 /**
  * @brief Enflame Management Library get device UUID info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevUuid(uint32_t dev_idx, char *p_info);

 /**
  * @brief Enflame Management Library get device SN(Serial Number) info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevSn(uint32_t dev_idx, char *p_sn);

 /**
  * @brief Enflame Management Library get device PN(Part Number) info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevPn(uint32_t dev_idx, char *p_pn);

 /**
  * @brief Enflame Management Library get device Manufacturing Date.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevMfd(uint32_t dev_idx, char *p_date);

 /**
  * @brief Enflame Management Library get device SKU info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevSKU(uint32_t dev_idx, char *p_info);

 /**
  * @brief Enflame Management Library get device hw arch info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetHwArch(uint32_t dev_idx, efmlHwArch_t *p_arch);

 /**
  * @brief Enflame Management Library get device hw arch name.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetHwArchName(uint32_t dev_idx, char *arch_name);

 /**
  * @brief Enflame Management Library get power cable qualify status.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetCableQualify(
     uint32_t dev_idx, efmlCableQualifySts_t *p_sts);

 /**
  * @brief Enflame Management Library get device PCIe slot or OAM name.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevSlotOamName(uint32_t dev_idx, char *p_slot_oam);

 /**
  * @brief Enflame Management Library get current device pcie link speed.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPcieLinkSpeed(uint32_t dev_idx, efmlPcieSpeed_t *p_link_speed);

 /**
  * @brief Enflame Management Library get current device pcie link width.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPcieLinkWidth(uint32_t dev_idx, efmlPcieWidth_t *p_link_width);

 /**
  * @brief Enflame Management Library get current device pcie link info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPcieLinkInfo(uint32_t dev_idx, efmlPcieLinkInfo_t *p_link_info);

 /**
  * @brief Enflame Management Library get pcie throughput info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPcieThroughput(uint32_t dev_idx, efmlPcieThroughputInfo_t *p_info);

 /**
  * @brief Enflame Management Library pcie hot reset.
  *
  */
EFML_EXPORT efmlReturn_t EfmlPcieHotReset(uint32_t dev_idx);

 /**
  * @brief Enflame Management Library advance pcie hot reset.
  *
  */
EFML_EXPORT efmlReturn_t EfmlPcieHotResetV2(uint32_t dev_idx, bool is_force);

 /**
  * @brief Enflame Management Library system pcie reset.
  *
  */
EFML_EXPORT efmlReturn_t EfmlPcieHotResetV3(uint32_t dev_idx);

 /**
  * @brief Enflame Management Library advance pcie function level reset.
  *
  */
EFML_EXPORT efmlReturn_t EfmlPcieFLR(uint32_t dev_idx, bool is_force);

 /**
  * @brief Enflame Management Library get pcie slot ID.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPciePhysicalSlotID(uint32_t dev_idx, uint32_t *p_id);

 /**
  * @brief Enflame Management Library get dram ECC status.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetEccStatus(uint32_t dev_idx, uint32_t *status);

 /**
  * @brief Enflame Management Library get total GCU-LARE port numbers.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLarePortNum(uint32_t dev_idx, uint32_t *p_data);

 /**
  * @brief Enflame Management Library get GCU-LARE port info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLarePortInfo(uint32_t dev_idx, uint32_t port_id, efmlGcuLarePortInfo_t *p_info);

 /**
  * @brief Enflame Management Library get GCU-LARE link info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLareLinkInfo(uint32_t dev_idx, uint32_t port_id, efmlGcuLareLinkInfo_t *p_info);

 /**
  * @brief Enflame Management Library get GCU-LARE dtuid info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLareId(uint32_t dev_idx, uint32_t *p_data);

 /**
  * @brief Enflame Management Library get GCU-LARE support info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLareIsSupported(uint32_t dev_idx, bool *is_gcu_lare_supported);

 /**
  * @brief Enflame Management Library get GCU-LARE throughput info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuLareThroughput(
     uint32_t dev_idx, uint32_t port_id, efmlGcuLareThroughputInfo_t *p_info);

 /**
  * @brief Enflame Management Library get the total number of special device's pgs.
  *
  */
 EFML_DEPRECATED_EXPORT efmlReturn_t EfmlGetPGCount(uint32_t dev_idx, uint32_t *pg_count);

 /**
  * @brief Enflame Management Library get the device pg usage.
  *
  */
 EFML_DEPRECATED_EXPORT efmlReturn_t EfmlGetDevPGUsage(uint32_t dev_idx, uint32_t pg_idx, float *p_data);

 /**
  * @brief Enflame Management Library get the device pg usage from background sample thread..
  *
  */
 EFML_DEPRECATED_EXPORT efmlReturn_t EfmlGetDevPGUsageAsync(uint32_t dev_idx, uint32_t pg_idx, float *p_data);

 /**
  * @brief Enflame Management Library get the total number of special device's clusters.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetClusterCount(uint32_t dev_idx, uint32_t *cluster_count);

 /**
  * @brief Enflame Management Library get the device cluster usage.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevClusterUsage(uint32_t dev_idx, uint32_t cluster_idx, float *p_data);

 /**
  * @brief Enflame Management Library get the cluster mem info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevClusterHbmMem(
     uint32_t dev_idx, uint32_t cluster_idx, efmlClusterHbmMemInfo_t *p_mem);

 /**
  * @brief Enflame Management Library get the device health stauts.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevHealth(uint32_t dev_idx, bool *health);

 /**
  * @brief Enflame Management Library get the device ecc stauts.
  * @deprecated ECC error counts in this API are not reliable, please use EfmlGetDevRmaDetailsV2 instead.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevEccStatus(uint32_t dev_idx, efmlEccStatus_t *p_status);

 /**
  * @brief Enflame Management Library get the device rma stauts.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevRmaStatus(uint32_t dev_idx, efmlRmaStatus_t *p_status);

 /**
  * @brief Enflame Management Library get the device rma details.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevRmaDetails(uint32_t dev_idx, efmlRmaDetails_t *p_details);

 /**
  * @brief Enflame Management Library get the device rma details.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevRmaDetailsV2(uint32_t dev_idx, efmlRmaDetailsV2_t *p_details);

 /**
  * @brief Enflame Management Library get the device logic id.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevLogicId(uint32_t dev_idx, uint32_t *p_logic_id);

 /**
  * @brief Enflame Management Library get the device decoder capability.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetDevDecoderCap(uint32_t dev_idx, efmlDecoderCap_t *p_decoder_cap);

 /**
  * @brief Enflame Management Library set the device ecc mode.
  *
  */
EFML_EXPORT efmlReturn_t EfmlSetDevEccMode(uint32_t dev_idx, bool enable);

 /**
  * @brief Enflame Management Library set the device hbm scan mode.
  *
  */
EFML_EXPORT efmlReturn_t EfmlHbmScanMode(uint32_t dev_idx, efmlHbmScanType_t op_type);

 /**
  * @brief Enflame Management Library get event message.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetEvent(int timeout_ms, efmlEvent_t *p_event);

 /**
  * @brief Enflame Management Library get event message. Non-blocking.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetEventV2(uint32_t dev_idx, efmlEvent_t *p_event);

 /**
  * @brief Enflame Management Library start listen event.
  *
  */
 EFML_DEPRECATED_EXPORT efmlReturn_t EfmlStartListenEvent(uint32_t dev_idx);

 /**
  * @brief Enflame Management Library switch operate space.
  *
  */
EFML_EXPORT efmlReturn_t EfmlSwitchOperateSpace(efmlOperateSpace_t opsp);

 /**
  * @brief Enflame Management Library get process info on device.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetProcessInfo(
     uint32_t dev_idx, uint32_t *process_count, efmlProcessInfo_t *p_info);

 /**
  * @brief Enflame Management Library get device numa node.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetNumaNode(uint32_t dev_idx, int *numa_node);

 /**
  * @brief Enflame Management Library get PF name of device. Only support on VF.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetSriovPfName(uint32_t dev_idx, char *pf_name);

 /**
  * @brief Enflame Management Library get device affinity CPU list.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetAffinityCpuList(uint32_t dev_idx, char *cpu_list);

 /**
  * @brief Enflame Management Library get SR-IOV status.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetGcuVirtStatus(uint32_t dev_idx, efmlGcuVirtMode_t *virt_mode);

 /**
  * @brief Enflame Management Library get SR-IOV VF number. Only support on PF.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetSriovVfNum(uint32_t dev_idx, uint32_t *num);


 /**
  * @brief Enflame Management Library get VPU usage.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetVpuUsage(uint32_t dev_idx, efmlVpuUsage_t *p_usage);

 /**
  * @brief Enflame Management Library get multi chip modules mode.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetMcmMode(uint32_t dev_idx, efmlMcmMode_t *mcm_mode);

 /**
  * @brief Enflame Management Library set multi chip modules mode. Only support on PF.
  *
  */
EFML_EXPORT efmlReturn_t EfmlSetMcmMode(uint32_t dev_idx, efmlMcmMode_t mcm_mode);


 /**
  * @brief Enflame Management Library get power mode.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetPowerMode(uint32_t dev_idx, efmlPowerMode_t *power_mode);


 /**
  * @brief Enflame Management Library set power mode.
  *
  */
EFML_EXPORT efmlReturn_t EfmlSetPowerMode(uint32_t dev_idx, efmlPowerMode_t power_mode);


 /**
  * @brief Enflame Management Library get RAS info.
  *
  */
EFML_EXPORT efmlReturn_t EfmlGetErrorRecords(uint32_t dev_idx, uint32_t count, efmlErrorRecord_t *p_records,
     uint32_t *total_count);

/**
 * @brief Enflame Management Library initialization with log file name
 *
 * @return efmlReturn_t
 */
efmlReturn_t EfmlInitFile(char *log_file_name, bool no_driver);

/**
 * @brief Enflame Management Library get the device temperature.
 *
 */
efmlReturn_t EfmlGetDevTemp(uint32_t dev_idx, efmlDevThermalInfo_t* p_temp);

/**
 * @brief Enflame Management Library get the device DPM level.
 *
 */
efmlReturn_t EfmlGetDevDpmLevel(uint32_t dev_idx, uint32_t* p_dpm);

/**
 * @brief Enflame Management Library switch dtu power stock mode.
 *
 */
efmlReturn_t EfmlSetDevSupportPowerStock(uint32_t dev_idx, bool enable_power_stock_support);

/**
 * @brief Enflame Management Library get if dtu support power stock mode.
 *
 */
efmlReturn_t EfmlGetDevSupportPowerStock(uint32_t dev_idx, bool* is_support_power_stock);

/**
 * @brief Enflame Management Library dump supported device list.
 *
 */
efmlReturn_t EfmlDumpDevList(void);

/**
 * @brief Enflame Management Library get device parent info.
 *
 */
efmlReturn_t EfmlGetDevParentInfo(uint32_t dev_idx, efmlDeviceInfo_t *p_info);


/**
 * @brief Enflame Management Library get device PCIe slot number.
 *
 */
efmlReturn_t EfmlGetDevSlotNum(uint32_t dev_idx, uint32_t *p_slot);

/**
 * @brief Enflame Management Library select one target device by index.
 *
 */
efmlReturn_t EfmlSelDevByIndex(uint32_t dev_idx);

/**
 * @brief Enflame Management Library get pcie up top.
 *
 */
efmlReturn_t EfmlGetDevTop(uint32_t dev_idx);

/**
 * @brief Enflame Management Library get virtualization status.
 *
 */
efmlReturn_t EfmlGetDevIsVdtuEnabled(uint32_t dev_idx, bool *is_vdtu_enabled);

/**
 * @brief Enflame Management Library get if dtu support low power mode.
 *
 */
efmlReturn_t EfmlGetDevSupportLowPower(uint32_t dev_idx, bool *is_support_low_power);

/**
 * @brief Enflame Management Library switch dtu low power mode.
 *
 */
efmlReturn_t EfmlSetDevSupportLowPower(uint32_t dev_idx, bool enable_low_power_support);

/**
 * @brief Enflame Management Library get total esl port numbers.
 *
 */
efmlReturn_t EfmlGetEslPortNum(uint32_t dev_idx, uint32_t *p_data);

/**
 * @brief Enflame Management Library get esl port info.
 *
 */
efmlReturn_t EfmlGetEslPortInfo(uint32_t dev_idx, uint32_t port_id, efmlEslPortInfo_t *p_info);

/**
 * @brief Enflame Management Library get esl link info.
 *
 */
efmlReturn_t EfmlGetEslLinkInfo(uint32_t dev_idx, uint32_t port_id, efmlEslLinkInfo_t *p_info);

/**
 * @brief Enflame Management Library get esl dtuid info.
 *
 */
efmlReturn_t EfmlGetEslDtuId(uint32_t dev_idx, uint32_t *p_data);

/**
 * @brief Enflame Management Library get esl support info.
 *
 */
efmlReturn_t EfmlGetEslIsSupported(uint32_t dev_idx, bool *is_esl_supported);

/**
 * @brief Enflame Management Library get esl throughput info.
 *
 */
efmlReturn_t EfmlGetEslThroughput(
    uint32_t dev_idx, uint32_t port_id, efmlEslThroughputInfo_t *p_info);

#ifdef __cplusplus
} // extern "C"
#endif


#endif  //__EFML_H_
