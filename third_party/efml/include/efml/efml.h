/////////////////////////////////////////////////////////////////////////////
//
//  @file efml.h
//
//  @brief API interface of Enflame Management Library
//
//  Enflame Tech, All Rights Reserved. 2024-2026 Copyright (C)
//
/////////////////////////////////////////////////////////////////////////////

#ifndef LIB_EFML_INCLUDE_EFML_H_
#define LIB_EFML_INCLUDE_EFML_H_

#include <stdbool.h>
#include <stdint.h>

#include "efml/efml_export.h"

#define KiB (1024)
#define MiB (1024 * KiB)
#define GiB (1024 * MiB)

#ifdef __cplusplus
extern "C" {
#endif

#define MAX_DRS_INSTANCE_COUNT (24)

#define EFML_STRUCT_VERSION(data, ver) (unsigned int)(sizeof(efml ## data ## _v ## ver ## _t) | \
                                                      (ver << 24U))

/***************************************************************************************************/
/** @defgroup efmlDeviceStructs Device Structs
 *  @{
 */
/***************************************************************************************************/

typedef struct efmlDevice_st* efmlDevice_t;

/**
 * Buffer size guaranteed to be large enough for pci bus id
 */
#define EFML_DEVICE_PCI_BUS_ID_BUFFER_SIZE      32

#define EFML_DEVICE_PCI_BUS_ID_BUFFER_V2_SIZE   16

/**
 * PCI information about a GCU device.
 */
typedef struct efmlPciInfo_st
{
    char busIdLegacy[EFML_DEVICE_PCI_BUS_ID_BUFFER_V2_SIZE];  //!< The legacy tuple domain:bus:device.function PCI identifier (&amp; NULL terminator)
    unsigned int domain;                                      //!< The PCI domain on which the device's bus resides, 0 to 0xffffffff
    unsigned int bus;                                         //!< The bus on which the device resides, 0 to 0xff
    unsigned int device;                                      //!< The device's id on the bus, 0 to 31
    unsigned int function;                                    //!< The device's function number, 0 to 7
    unsigned int pciDeviceId;                                 //!< The combined 16-bit device id and 16-bit vendor id
    unsigned int pciSubSystemId;                              //!< The 32-bit Sub System Device ID
    char busId[EFML_DEVICE_PCI_BUS_ID_BUFFER_SIZE];           //!< The tuple domain:bus:device.function PCI identifier (&amp; NULL terminator)
} efmlPciInfo_t;

/**
 * PCI format string for \p busIdLegacy
 */
#define EFML_DEVICE_PCI_BUS_ID_LEGACY_FMT           "%04x:%02x:%02x.%x"

/**
 * PCI format string for \p busId
 */
#define EFML_DEVICE_PCI_BUS_ID_FMT                  "%04x:%02x:%02x.%x"

/**
 * Utility macro for filling the pci bus id format from a efmlPciInfo_t
 */
#define EFML_DEVICE_PCI_BUS_ID_FMT_ARGS(pciInfo)    (pciInfo)->domain, \
                                                     (pciInfo)->bus,    \
                                                     (pciInfo)->device, \
                                                     (pciInfo)->function

/**
 * Utilization information for a device.
 */
typedef struct efmlUtilization_st
{
    unsigned int gcu;                //!< Percent of time over the past sample period during which one or more kernels was executing on the GCU
    unsigned int memory;             //!< Percent of time over the past sample period during which global (device) memory was being read or written
} efmlUtilization_t;

/**
 * Occupancy for the device's execution units, as returned by \ref efmlDeviceGetOccupancy.
 * Values are percentages in the range 0~100, analogous to \ref efmlUtilization_t.
 */
typedef struct efmlOccupancy_st
{
    unsigned int sip;                  //!< SIP occupancy: busy SIPs as a percent of total SIPs in the sample
    unsigned int sip_vector;           //!< Reserved; not populated by the current driver
    unsigned int sip_matrix;           //!< Reserved; not populated by the current driver
    unsigned int sip_special_function; //!< Reserved; not populated by the current driver
    unsigned int memory;               //!< Reserved; not populated by the current driver
} efmlOccupancy_t;

/**
 * Memory allocation information for a device.
 */
typedef struct efmlMemory_st
{
    unsigned long long total;        //!< Total installed device memory (in bytes)
    unsigned long long free;         //!< Unallocated device memory (in bytes)
    unsigned long long used;         //!< Allocated device memory (in bytes).
                                     //!< Note that the driver/GCU always sets aside a small amount of memory for bookkeeping
} efmlMemory_t;

/**
 * Memory allocation information for a device (v2).
 */
typedef struct efmlMemory_v2_st
{
    unsigned int version;            //!< Structure format version (must be 2)
    unsigned long long total;        //!< Total installed device memory (in bytes)
    unsigned long long reserved;     //!< Device memory (in bytes) reserved for system use (driver or firmware).
    unsigned long long free;         //!< Unallocated device memory (in bytes)
    unsigned long long used;         //!< Allocated device memory (in bytes).
} efmlMemory_v2_t;

#define efmlMemory_v2 EFML_STRUCT_VERSION(Memory, 2)

/**
 * BAR1 Memory Information for a device
 */
typedef struct efmlBAR1Memory_st
{
    unsigned long long bar1Total;    //!< Total BAR1 Memory (in bytes)
    unsigned long long reserved1;    //!< Reserved
    unsigned long long reserved2;    //!< Reserved
} efmlBAR1Memory_t;

/**
 * BAR2 Memory Information for a device
 */
typedef struct efmlBAR2Memory_st
{
    unsigned long long bar2Total;    //!< Total BAR2 Memory (in bytes)
    unsigned long long reserved1;    //!< Reserved
    unsigned long long reserved2;    //!< Reserved
} efmlBAR2Memory_t;

/**
 * Information about running compute processes on the GCU
 */
typedef struct efmlProcessInfo_st
{
    unsigned int pid;                 //!< Process ID
    unsigned long long usedGcuMemory; //!< Amount of used GCU memory in bytes.
    unsigned int gcuInstanceId;       //!< always 0xFFFFFFFF
    unsigned int computeInstanceId;   //!< always 0xFFFFFFFF
} efmlProcessInfo_v1_t;

typedef struct efmlGcuLareInfo_st {
    unsigned int maxTthroughput;       //!< The max throughput in Gbps
    unsigned long long mac[8];         //!< The MAC address of the GCU-LARE
    unsigned int ip[8];                //!< The IP address of the GCU-LARE
    unsigned int ip_mask_prefix[8];    //!< The IP mask prefix of the GCU-LARE
    unsigned int gateway;              //!< The gateway of the GCU-LARE
    unsigned char remoteUuid[16];      //!< The UUID of the remote GCU-LARE
    char nicName[16];                  //!< The NIC name of the GCU-LARE
    unsigned long long throughput;
} efmlGcuLareInfo_t;

/**
 * GCU-LARE port QoS / PFC-related parameters from the device.
 * Used with \ref efmlDeviceGetGcuLareQosParam and \ref efmlDeviceSetGcuLareQosParam.
 */
typedef struct efmlGcuLareQosParam_st {
    unsigned int dscp[4];                //!< DSCP class selector per traffic index (four entries per port configuration)
    unsigned int priority[4];            //!< Priority level per traffic index (four entries; device-specific mapping)
    unsigned int xon;                    //!< PFC XON threshold (resume transmit when buffer occupancy drops below this)
    unsigned int xoff;                   //!< PFC XOFF threshold (request pause when buffer occupancy reaches this)
} efmlGcuLareQosParam_t;

typedef struct efmlGcuLareTransactionConfig_st {
    unsigned long long timeout;  //!< The timeout for the GCU-LARE transaction in nanoseconds
    unsigned int retry_cnt;      //!< The retry count for the GCU-LARE transaction
} efmlGcuLareTransactionConfig_t;

#define EFML_DEVICE_GCU_LARE_QUEUE_SIZE 256 //!< The max count of the GCU-LARE queue in one port

typedef struct efmlGcuLareQueueDetail_st {
    unsigned int status;                //!< The status of the queue
    unsigned int sourceIp;              //!< The source IP address of the queue
    unsigned int destinationIp;         //!< The destination IP address of the queue
    unsigned long long sourceMac;       //!< The source MAC address of the queue
    unsigned long long destinationMac;  //!< The destination MAC address of the queue
    unsigned int destinationQueueId;    //!< The destination queue ID of the queue
    unsigned int sourceUdpPort;         //!< The source UDP port of the queue
    unsigned int pfcBuffer;             //!< The PFC buffer of the queue
} efmlGcuLareQueueDetail_t;

/**
 * Enum to represent GcuLare queryable capabilities
 */
typedef enum efmlGcuLareCapability_enum
{
    EFML_GCULARE_CAP_P2P_SUPPORTED = 0,     //!< P2P over GcuLare is supported
    EFML_GCULARE_CAP_SYSMEM_ACCESS = 1,     //!< Access to system memory is supported
    EFML_GCULARE_CAP_P2P_ATOMICS   = 2,     //!< P2P atomics are supported
    EFML_GCULARE_CAP_SYSMEM_ATOMICS= 3,     //!< System memory atomics are supported
    EFML_GCULARE_CAP_SLI_BRIDGE    = 4,     //!< SLI is supported over this link
    EFML_GCULARE_CAP_VALID         = 5,     //!< Link is supported on this device
    EFML_GCULARE_CAP_COUNT
} efmlGcuLareCapability_t;

/**
 * Enum to represent GcuLare's remote device type
 */
typedef enum efmlIntGcuLareDeviceType_enum
{
    EFML_GCULARE_DEVICE_TYPE_GCU     = 0x00, //!< The GCU device type

    EFML_GCULARE_DEVICE_TYPE_SWITCH  = 0x02, //!< The switch device type
    EFML_GCULARE_DEVICE_TYPE_UNKNOWN = 0xFF, //!< The unknown device type
} efmlIntGcuLareDeviceType_t;

/**
 * Represents Type of Sampling Event
 */
typedef enum efmlSamplingType_enum
{
    EFML_TOTAL_POWER_SAMPLES        = 0, //!< To represent total power drawn by GCU
    EFML_GCU_UTILIZATION_SAMPLES    = 1, //!< To represent percent of time during which one or more kernels was executing on the GCU
    EFML_MEMORY_UTILIZATION_SAMPLES = 2, //!< To represent percent of time during which global (device) memory was being read or written
    EFML_ENC_UTILIZATION_SAMPLES    = 3, //!< To represent percent of time during which encoder remains busy
    EFML_DEC_UTILIZATION_SAMPLES    = 4, //!< To represent percent of time during which decoder remains busy
    EFML_PROCESSOR_CLK_SAMPLES      = 5, //!< To represent processor clock samples
    EFML_MEMORY_CLK_SAMPLES         = 6, //!< To represent memory clock samples
    // Keep this last
    EFML_SAMPLINGTYPE_COUNT
} efmlSamplingType_t;

/**
 * Represents the queryable PCIe utilization counters
 */
typedef enum efmlPcieUtilCounter_enum
{
    EFML_PCIE_UTIL_TX_BYTES             = 0, //!< 1KB granularity
    EFML_PCIE_UTIL_RX_BYTES             = 1, //!< 1KB granularity
    EFML_PCIE_UTIL_COUNT
} efmlPcieUtilCounter_t;

/**
 * Scalar type tag for a field value sample (\ref efmlFieldValue_t::valueType).
 * The active member of \ref efmlValue_t is determined by this enumerator.
 */
typedef enum efmlValueType_enum
{
    EFML_VALUE_TYPE_DOUBLE = 0,               //!< \ref efmlValue_t::dVal
    EFML_VALUE_TYPE_UNSIGNED_INT = 1,         //!< \ref efmlValue_t::uiVal
    EFML_VALUE_TYPE_UNSIGNED_LONG = 2,        //!< \ref efmlValue_t::ulVal
    EFML_VALUE_TYPE_UNSIGNED_LONG_LONG = 3,   //!< \ref efmlValue_t::ullVal
    EFML_VALUE_TYPE_SIGNED_LONG_LONG = 4,     //!< \ref efmlValue_t::sllVal
    EFML_VALUE_TYPE_SIGNED_INT = 5,           //!< \ref efmlValue_t::siVal
    EFML_VALUE_TYPE_UNSIGNED_SHORT = 6,      //!< \ref efmlValue_t::usVal

    // Keep this last
    EFML_VALUE_TYPE_COUNT                    //!< Number of scalar types; not a valid \c valueType for samples
} efmlValueType_t;

/**
 * Union to represent different types of Value
 */
typedef union efmlValue_st
{
    double dVal;                    //!< If the value is double
    int siVal;                      //!< If the value is signed int
    unsigned int uiVal;             //!< If the value is unsigned int
    unsigned long ulVal;            //!< If the value is unsigned long
    unsigned long long ullVal;      //!< If the value is unsigned long long
    signed long long sllVal;        //!< If the value is signed long long
    unsigned short usVal;           //!< If the value is unsigned short
} efmlValue_t;

/**
 * Information for Sample
 */
typedef struct efmlSample_st
{
    unsigned long long timeStamp;       //!< CCU Timestamp in microseconds
    efmlValue_t sampleValue;            //!< Sample Value
} efmlSample_t;


/** @} */

/***************************************************************************************************/
/** @defgroup efmlDeviceEnumvs Device Enums
 *  @{
 */
/***************************************************************************************************/

/**
 * Generic enable/disable enum.
 */
typedef enum efmlEnableState_enum
{
    EFML_FEATURE_DISABLED    = 0,     //!< Feature disabled
    EFML_FEATURE_ENABLED     = 1      //!< Feature enabled
} efmlEnableState_t;

/**
 * Temperature thresholds.
 */
typedef enum efmlTemperatureThresholds_enum
{
    EFML_TEMPERATURE_THRESHOLD_SHUTDOWN = 0,    //!< Temperature at which the GCU will shut down for HW protection
    EFML_TEMPERATURE_THRESHOLD_SLOWDOWN = 1,    //!< Temperature at which the GCU will slowdown to the minimum clock
    EFML_TEMPERATURE_THRESHOLD_MEM_MAX  = 2,    //!< Max temperature at which the memory works properly
    EFML_TEMPERATURE_THRESHOLD_GCU_MAX  = 3,    //!< Max temperature at which the GCU core works properly
    //!< Keep this last
    EFML_TEMPERATURE_THRESHOLD_COUNT
} efmlTemperatureThresholds_t;

/**
 * Temperature sensors.
 */
typedef enum efmlTemperatureSensors_enum
{
    EFML_TEMPERATURE_GCU      = 0,    //!< Temperature sensor for the GCU die
    EFML_TEMPERATURE_MEMORY   = 1,    //!< Temperature sensor for the Memory
    EFML_TEMPERATURE_BOARD    = 2,    //!< Temperature sensor for the Board

    // Keep this last
    EFML_TEMPERATURE_COUNT
} efmlTemperatureSensors_t;

/**
 * Memory error types
 */
typedef enum efmlMemoryErrorType_enum
{
    /**
    * A memory error that was corrected
    *
    * For ECC errors, these are single bit errors
    */
    EFML_MEMORY_ERROR_TYPE_CORRECTED = 0,
    /**
    * A memory error that was not corrected
    *
    * For ECC errors, these are double bit errors
    */
    EFML_MEMORY_ERROR_TYPE_UNCORRECTED = 1,

    // Keep this last
    EFML_MEMORY_ERROR_TYPE_COUNT //!< Count of memory error types

} efmlMemoryErrorType_t;

/**
 * ECC counter types.
 *
 * Note: Volatile counts are reset each time the driver loads.
 */
typedef enum efmlEccCounterType_enum
{
    EFML_VOLATILE_ECC      = 0,      //!< Volatile counts are reset each time the driver loads.
    EFML_AGGREGATE_ECC     = 1,      //!< Aggregate counts persist across reboots (i.e. for the lifetime of the device)

    // Keep this last
    EFML_ECC_COUNTER_TYPE_COUNT      //!< Count of memory counter types
} efmlEccCounterType_t;

/**
 * Clock types.
 *
 * All speeds are in Mhz.
 */
typedef enum efmlClockType_enum
{
    EFML_CLOCK_GRAPHICS  = 0,        //!< Graphics clock domain
    EFML_CLOCK_SIP       = 1,        //!< SIP clock domain
    EFML_CLOCK_MEM       = 2,        //!< Memory clock domain
    EFML_CLOCK_VIDEO     = 3,        //!< Video encoder/decoder clock domain

    // Keep this last
    EFML_CLOCK_COUNT //!< Count of clock types
} efmlClockType_t;


#define EFML_MAX_GCU_LOW_POWER_MODES 4

/**
 * Allowed LPModes.
 */
typedef enum efmlLPModes_enum
{
    EFML_LP_0               = 1,       //!< Low Power mode 0 -- Maximum Performance
    EFML_LP_1               = 7,       //!< Low Power mode 1
    EFML_LP_2               = 5,       //!< Low Power mode 2
    EFML_LP_3               = 9,       //!< Low Power mode 3 -- Minimum Power Consumption
    EFML_LP_UNKNOWN         = 32       //!< Unknown Low Power mode
} efmlLPModes_t;


/**
 * Return values for EFML API calls.
 */
typedef enum efmlReturn_enum
{
    // cppcheck-suppress *
    EFML_SUCCESS_V2 = 0,                       //!< The operation was successful
    EFML_ERROR_UNINITIALIZED_V2 = 1,           //!< EFML was not first initialized with efmlInit()
    EFML_ERROR_INVALID_ARGUMENT_V2 = 2,        //!< A supplied argument is invalid
    EFML_ERROR_NOT_SUPPORTED_V2 = 3,           //!< The requested operation is not available on target device
    EFML_ERROR_NO_PERMISSION = 4,              //!< The current user does not have permission for operation
    EFML_ERROR_NOT_FOUND = 6,                  //!< A query to find an object was unsuccessful
    EFML_ERROR_INSUFFICIENT_SIZE = 7,          //!< An input argument is not large enough
    EFML_ERROR_INSUFFICIENT_POWER = 8,         //!< A device's external power cables are not properly attached
    EFML_ERROR_DRIVER_NOT_LOADED_V2 = 9,       //!< ENFLAME driver is not loaded
    EFML_ERROR_TIMEOUT_V2 = 10,                //!< User provided timeout passed
    EFML_ERROR_IRQ_ISSUE = 11,                 //!< ENFLAME Kernel detected an interrupt issue with a GCU
    EFML_ERROR_LIBRARY_NOT_FOUND_V2 = 12,      //!< EFML Shared Library couldn't be found or loaded
    EFML_ERROR_FUNCTION_NOT_FOUND_V2 = 13,     //!< Local version of EFML doesn't implement this function
    EFML_ERROR_CORRUPTED_INFOROM = 14,         //!< infoROM is corrupted
    EFML_ERROR_GCU_IS_LOST = 15,               //!< The GCU has fallen off the bus or has otherwise become inaccessible
    EFML_ERROR_RESET_REQUIRED = 16,            //!< The GCU requires a reset before it can be used again
    EFML_ERROR_OPERATING_SYSTEM = 17,          //!< The GCU control device has been blocked by the operating system/cgroups
    EFML_ERROR_LIB_RM_VERSION_MISMATCH = 18,   //!< RM detects a driver/library version mismatch
    EFML_ERROR_IN_USE = 19,                    //!< An operation cannot be performed because the GCU is currently in use
    EFML_ERROR_MEMORY = 20,                    //!< Insufficient memory
    EFML_ERROR_NO_DATA = 21,                   //!< No data
    EFML_ERROR_VGCU_ECC_NOT_SUPPORTED = 22,    //!< The requested vGCU operation is not available on target device, because ECC is enabled
    EFML_ERROR_INSUFFICIENT_RESOURCES = 23,    //!< Ran out of critical resources, other than memory
    EFML_ERROR_FREQ_NOT_SUPPORTED = 24,        //!< Ran out of critical resources, other than memory
    EFML_ERROR_ARGUMENT_VERSION_MISMATCH = 25, //!< The provided version is invalid/unsupported
    EFML_ERROR_DEPRECATED  = 26,               //!< The requested functionality has been deprecated
    EFML_ERROR_NOT_READY = 27,                 //!< The system is not ready for the request
    EFML_ERROR_GCU_NOT_FOUND = 28,             //!< No GCUs were found
    EFML_ERROR_INVALID_STATE = 29,             //!< Resource not in correct state to perform requested operation
    EFML_ERROR_UNKNOWN = 999                   //!< An internal driver error occurred
} efmlReturn_v1_t;

/**
 * See \ref efmlDeviceGetMemoryErrorCounter
 */
typedef enum efmlMemoryLocation_enum
{
    EFML_MEMORY_LOCATION_L1_CACHE        = 0,    //!< GCU L1 Cache
    EFML_MEMORY_LOCATION_L2_CACHE        = 1,    //!< GCU L2 Cache
    EFML_MEMORY_LOCATION_DRAM            = 2,    //!< DRAM
    EFML_MEMORY_LOCATION_DEVICE_MEMORY   = 2,    //!< GCU Device Memory
    EFML_MEMORY_LOCATION_SRAM            = 7,    //!< SRAM
    // Keep this last
    EFML_MEMORY_LOCATION_COUNT              //!< This counts the number of memory locations the driver knows about
} efmlMemoryLocation_t;

/**
 * Structure to store utilization value and process Id
 */
typedef struct efmlProcessUtilizationSample_st
{
    unsigned int pid;                   //!< PID of process
    unsigned long long timeStamp;       //!< CPU Timestamp in microseconds
    unsigned int sipUtil;               //!< SIP Util Value
    unsigned int memUtil;               //!< Memory Util Value
    unsigned int encUtil;               //!< Encoder Util Value. Not support now
    unsigned int decUtil;               //!< Decoder Util Value. Not support now
} efmlProcessUtilizationSample_t;


/**
 * Simplified chip architecture
 */
#define EFML_DEVICE_ARCH_GCU300 30 // Devices based on the Enflame Scorpio architecture
#define EFML_DEVICE_ARCH_GCU400 40 // Devices based on the Enflame Libra architecture
#define EFML_DEVICE_ARCH_GCU410 41 // Devices based on the Enflame Libra architecture
#define EFML_DEVICE_ARCH_GCU450 45 // Devices based on the Enflame Libra architecture

#define EFML_DEVICE_ARCH_UNKNOWN   0xffffffff // Anything else, presumably something newer

typedef unsigned int efmlDeviceArchitecture_t;

typedef enum efmlMcmMode_enum {
    EFML_MCM_SINGLE_MODE = 1,
    EFML_MCM_DUAL_MODE   = 2,
} efmlMcmMode_t;

/** @} */

/***************************************************************************************************/
/** @addtogroup virtualGCU
 *  @{
 */
/***************************************************************************************************/
/** @defgroup efmlVirtualGcuEnums vGCU Enums
 *  @{
 */
/***************************************************************************************************/

/**
 * GCU virtualization mode types.
 */
typedef enum efmlGcuVirtualizationMode_enum {
    EFML_GCU_VIRTUALIZATION_MODE_NONE = 0,         //!< Represents Bare Metal GCU
    EFML_GCU_VIRTUALIZATION_MODE_PASSTHROUGH = 1,  //!< Device is associated with GCU-Passthorugh
    EFML_GCU_VIRTUALIZATION_MODE_VGCU = 2,         //!< Device is associated with vGCU.
    EFML_GCU_VIRTUALIZATION_MODE_MDEV = 3,         //!< Device is associated with MDEV.
    EFML_GCU_VIRTUALIZATION_MODE_SRIOV_PF = 4,     //!< Device is associated with SRIOV PF.
    EFML_GCU_VIRTUALIZATION_MODE_SRIOV_VF = 5,     //!< Device is associated with SRIOV VF.
} efmlGcuVirtualizationMode_t;

/**
 * Host GCU SR-IOV modes
 */
typedef enum efmlHostVgcuMode_enum
{
    EFML_HOST_VGCU_MODE_NON_SRIOV    = 0,     //!< Non SR-IOV mode
    EFML_HOST_VGCU_MODE_SRIOV        = 1      //!< SR-IOV mode
} efmlHostVgcuMode_t;

/** @} */
/** @} */

/***************************************************************************************************/
/** @defgroup efmlErrorRecord RAS Error Record
 *  @{
 */
/***************************************************************************************************/

/**
 * Reset action recorded for a RAS error entry.
 */
typedef enum efmlResetType_enum
{
    EFML_RESET_TYPE_NONE = 0, //!< No reset
    EFML_RESET_TYPE_HOT  = 1, //!< PCI Express Hot reset
    EFML_RESET_TYPE_FLR  = 2, //!< Function Level Reset (FLR)
    EFML_RESET_TYPE_BLR  = 3, //!< Block Level Reset (BLR)
    EFML_RESET_TYPE_QLR  = 4, //!< Queue Level Reset (QLR)
} efmlResetType_t;

/**
 * Origin category for the reset, derived from the RAS record master identifier.
 */
typedef enum efmlResetSrc_enum
{
    EFML_RESET_SRC_USER     = 0, //!< User triggered
    EFML_RESET_SRC_INTERNAL = 1, //!< Internal / unspecified driver classification
    EFML_RESET_SRC_SIP      = 2, //!< SIP-related
    EFML_RESET_SRC_BUS      = 3, //!< Bus-related
    EFML_RESET_SRC_FW       = 4, //!< Firmware
    EFML_RESET_SRC_DTE      = 5, //!< DTE
    EFML_RESET_SRC_DRAM_HBM = 6, //!< DRAM / HBM
    EFML_RESET_SRC_PCIE     = 7, //!< PCIe
    EFML_RESET_SRC_GCU_LARE = 8, //!< GCU-LARE
    EFML_RESET_SRC_UNKNOWN  = 999, //!< Unknown or unmapped source
} efmlResetSrc_t;

/**
 * One RAS error record returned by \ref efmlDeviceGetErrorRecords.
 */
typedef struct efmlErrorRecord_st
{
    unsigned long long event_id;     //!< RAS event id: high 32 bits \c master_id, low 32 bits \c cause_id
    efmlResetType_t    reset_type;   //!< Reset action for this record (RAS \c action)
    efmlResetSrc_t     source;       //!< Classified origin of the reset
    unsigned long long time_stamp;   //!< Timestamp from the RAS record
    unsigned int       bdf;          //!< PCI BDF of the device when the record was retrieved
} efmlErrorRecord_t;

/** @} */

/***************************************************************************************************/
/** @defgroup efmlFieldValueEnums Field Value Enums
 *  @{
 */
/***************************************************************************************************/

/**
 * Field Identifiers.
 *
 * All Identifiers pertain to a device. Each ID is only used once and is guaranteed never to change.
 */
#define EFML_FI_DEV_GCULARE_PORT_COUNT                91  //!< Number of GcuLare ports present on the device
#define EFML_FI_DEV_PCIE_COUNT_RECEIVER_ERROR         175 //!< PCIe Receiver Error Counter
#define EFML_FI_DEV_PCIE_COUNT_BAD_TLP                176 //!< PCIe Bad TLP Counter
#define EFML_FI_DEV_PCIE_COUNT_BAD_DLLP               178 //!< PCIe Bad DLLP Counter
#define EFML_FI_MAX                                   200

/**
 * Information for a Field Value Sample
 */
typedef struct efmlFieldValue_st
{
    unsigned int fieldId;       //!< ID of the EFML field to retrieve. This must be set before any call that uses this struct. See the constants starting with EFML_FI_ above.
    unsigned int scopeId;       //!< Scope ID can represent data used by EFML depending on fieldId's context. For example, for GcuLare throughput counter data, scopeId can represent portId.
    long long timestamp;        //!< CPU Timestamp of this value in microseconds since 1970
    long long latencyUsec;      //!< How long this field value took to update (in usec) within EFML. This may be averaged across several fields that are serviced by the same driver call.
    efmlValueType_t valueType;  //!< Type of the value stored in value
    efmlReturn_v1_t efmlReturn; //!< Return code for retrieving this value. This must be checked before looking at value, as value is undefined if efmlReturn != EFML_SUCCESS_V2
    efmlValue_t value;          //!< Value for this field. This is only valid if efmlReturn == EFML_SUCCESS_V2
} efmlFieldValue_t;

/** @} */

/***************************************************************************************************/
/** @addtogroup efmlEvents
 *  @{
 */
/***************************************************************************************************/

/**
 * Handle to an event set
 */
typedef struct efmlEventSet_st* efmlEventSet_t;

/** @defgroup efmlEventType Event Types
 *  Event Types which user can be notified about.
 *  See description of particular functions for details.
 *
 *  See \ref efmlDeviceRegisterEvents and \ref efmlDeviceGetSupportedEventTypes to check which devices
 *  support each event.
 *
 *  Types can be combined with bitwise or operator '|' when passed to \ref efmlDeviceRegisterEvents
 *  @{
 */
/***************************************************************************************************/

#define efmlEventTypeNone                       0x0000000000000000LL //!< Mask with no events
#define efmlEventTypeEidCriticalError           0x0000000000000008LL //!< Event that Eid critical error occurred
#define efmlEventTypeFatalPoisonError           0x0000000000002000LL //!< Event for Fatal Poison
#define efmlEventTypeGpuRecoveryAction          0x0000000000008000LL //!< Event for GPU Recovery Action

/** Mask of all events */
#define efmlEventTypeAll (efmlEventTypeNone    \
        | efmlEventTypeEidCriticalError        \
        | efmlEventTypeFatalPoisonError        \
        | efmlEventTypeGpuRecoveryAction)

/** @} */

/**
 * Information about occurred event
 */
typedef struct efmlEventData_st
{
    efmlDevice_t        device;             //!< Specific device where the event occurred
    unsigned long long  eventType;          //!< Information about what specific event occurred
    unsigned long long  eventData;          //!< Stores Eid error for the device in the event of efmlEventTypeEidCriticalError,
                                            //   eventData is 0 for any other event. eventData is set as 999 for unknown Eid error.
    unsigned int        gpuInstanceId;      //!< always set to 0xFFFFFFFF now
    unsigned int        computeInstanceId;  //!< always set to 0xFFFFFFFF now
} efmlEventData_t;

/** @} */

/***************************************************************************************************/
/** @defgroup efmlFabricDefs Fabric definitions
 *  @{
 */
/***************************************************************************************************/

#define EFML_GCU_FABRIC_UUID_LEN 16  //!< Length of Fabric UUID

/**
 * Fabric Probe States
 */
#define EFML_GCU_FABRIC_STATE_NOT_SUPPORTED 0  //!< Fabric Probe State not supported
#define EFML_GCU_FABRIC_STATE_NOT_STARTED   1  //!< Fabric Probe has not started
#define EFML_GCU_FABRIC_STATE_IN_PROGRESS   2  //!< Fabric Probe in progress
#define EFML_GCU_FABRIC_STATE_COMPLETED     3  //!< Fabric Probe State completed

/**
 * Probe State of GPU registration process
 */
typedef unsigned char efmlGcuFabricState_t;

typedef struct {
    unsigned char        clusterUuid[EFML_GCU_FABRIC_UUID_LEN]; //!< Uuid of the cluster to which this GCU belongs
    efmlReturn_v1_t      status;                                //!< Error status, if any. Must be checked only if state returns "complete".
    unsigned int         cliqueId;                              //!< ID of the fabric clique to which this GCU belongs
    efmlGcuFabricState_t state;                                 //!< Current state of GCU registration process
} efmlGcuFabricInfo_t;

/**
 * Fabric Degraded BW
 */
#define EFML_GCU_FABRIC_HEALTH_MASK_DEGRADED_BW_NOT_SUPPORTED 0  //!< Fabric Health Mask: Degraded Bandwidth not supported
#define EFML_GCU_FABRIC_HEALTH_MASK_DEGRADED_BW_TRUE          1  //!< Fabric Health Mask: Bandwidth degraded
#define EFML_GCU_FABRIC_HEALTH_MASK_DEGRADED_BW_FALSE         2  //!< Fabric Health Mask: Bandwidth not degraded

#define EFML_GCU_FABRIC_HEALTH_MASK_SHIFT_DEGRADED_BW 0          //!< Fabric Health Mask Bit Shift for Degraded Bandwidth
#define EFML_GCU_FABRIC_HEALTH_MASK_WIDTH_DEGRADED_BW 0x11       //!< Fabric Health Mask Width for Degraded Bandwidth

/**
 * GCU Fabric Health Status Mask for various fields can be obtained
 * using the below macro.
 * Ex - EFML_GCU_FABRIC_HEALTH_GET(var, _DEGRADED_BW)
 */
#define EFML_GCU_FABRIC_HEALTH_GET(var, type)             \
    (((var) >> EFML_GCU_FABRIC_HEALTH_MASK_SHIFT##type) & \
     (EFML_GCU_FABRIC_HEALTH_MASK_WIDTH##type))

/**
 * GPU Fabric Health Status Mask for various fields can be tested
 * using the below macro.
 * Ex - EFML_GCU_FABRIC_HEALTH_TEST(var, _DEGRADED_BW, _TRUE)
 */
#define EFML_GCU_FABRIC_HEALTH_TEST(var, type, val) \
    (EFML_GCU_FABRIC_HEALTH_GET(var, type) ==       \
     EFML_GCU_FABRIC_HEALTH_MASK##type##val)

/**
 * GCU Fabric information (v2).
 *
 * Version 2 adds the \ref efmlGcuFabricInfo_v2_t.version field
 * to the start of the structure, and the \ref efmlGcuFabricInfo_v2_t.healthMask
 * field to the end. This structure is not backwards-compatible with
 * \ref efmlGcuFabricInfo_t.
 */
typedef struct {
    unsigned int         version;                               //!< Structure version identifier (set to \ref efmlGcuFabricInfo_v2)
    unsigned char        clusterUuid[EFML_GCU_FABRIC_UUID_LEN]; //!< Uuid of the cluster to which this GCU belongs
    efmlReturn_v1_t      status;                                //!< Error status, if any. Must be checked only if state returns "complete".
    unsigned int         cliqueId;                              //!< ID of the fabric clique to which this GCU belongs
    efmlGcuFabricState_t state;                                 //!< Current state of GCU registration process
    unsigned int         healthMask;                            //!< GCU Fabric health Status Mask
} efmlGcuFabricInfo_v2_t;

typedef efmlGcuFabricInfo_v2_t efmlGcuFabricInfoV_t;

/**
 * Version identifier value for \ref efmlGcuFabricInfo_v2_t.version.
 */
#define efmlGcuFabricInfo_v2 EFML_STRUCT_VERSION(GcuFabricInfo, 2)

/** @} */

/***************************************************************************************************/
/** @defgroup efmlInitializationAndCleanup Initialization and Cleanup
 * This chapter describes the methods that handle EFML initialization and cleanup.
 * It is the user's responsibility to call \ref efmlInit_v2() before calling any other methods, and
 * efmlShutdown() once EFML is no longer being used.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Initialize the EFML library.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @note This function must be called before any other EFML functions.
 */
EFML_EXPORT efmlReturn_v1_t efmlInit(void);
EFML_EXPORT efmlReturn_v1_t efmlInit_v2(void);

/**
 * @brief Initialize the EFML library for physical functions.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @note This function must be called before any other EFML functions.
 * @note Only call this function if SR-IOV is enabled.
 * @note This function is conflict with efmlInit(). If efmlInit() is called, efmlPfInit() will do nothing.
 */
EFML_EXPORT efmlReturn_v1_t efmlPfInit(void);

/**
 * @brief Shutdown the EFML library.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @note This function must be called to release resources allocated by efmlInit().
 *      After calling this function, no other EFML functions should be called.
 */
EFML_EXPORT efmlReturn_v1_t efmlShutdown(void);

/**
 * @brief Shutdown the EFML library for physical functions.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @note This function must be called to release resources allocated by efmlInit().
 *      After calling this function, no other EFML functions should be called.
 * @note Only call this function after efmlPfInit().
 */
EFML_EXPORT efmlReturn_v1_t efmlPfShutdown(void);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlErrorReporting Error reporting
 * This chapter describes helper functions for error reporting routines.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Helper method for converting EFML error codes to strings.
 *
 * @param error [in] The error code to convert.
 *
 * @return A string representation of the error code.
 */
EFML_EXPORT const char* efmlErrorString(efmlReturn_v1_t error);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlConstants Constants
 *  @{
 */
/***************************************************************************************************/

/**
 * Buffer size guaranteed to be large enough for storing GCU identifiers.
 */
#define EFML_DEVICE_UUID_BUFFER_SIZE                  80

/**
 * Buffer size guaranteed to be large enough for \ref efmlDeviceGetBoardPartNumber
 */
#define EFML_DEVICE_PART_NUMBER_BUFFER_SIZE           80

/**
 * Buffer size guaranteed to be large enough for \ref efmlSystemGetDriverVersion
 */
#define EFML_SYSTEM_DRIVER_VERSION_BUFFER_SIZE        80

/**
 * Buffer size guaranteed to be large enough for \ref efmlDeviceGetName
 */
#define EFML_DEVICE_NAME_V2_BUFFER_SIZE               96

/**
 * Buffer size guaranteed to be large enough for \ref efmlDeviceGetSerial
 */
#define EFML_DEVICE_SERIAL_BUFFER_SIZE                30

/**
 * Buffer size guaranteed to be large enough for \ref efmlDeviceGetVbiosVersion
 */
#define EFML_DEVICE_VBIOS_VERSION_BUFFER_SIZE         32

/** @} */

/***************************************************************************************************/
/** @defgroup efmlSystemQueries System Queries
 * This chapter describes the queries that EFML can perform against the local system. These queries
 * are not device-specific.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Retrieves the version of the system's GCU driver.
 *
 * @param version [out] Reference where the version string is returned.
 * @param length [in] The length of the version string buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlSystemGetDriverVersion(char* version, unsigned int length);

/**
 * @brief Retrieves the version of EFML library.
 *
 * @param version [out] Reference where the version string is returned.
 * @param length [in] The length of the version string buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlSystemGetEFMLVersion(char* version, unsigned int length);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlDeviceQueries Device Queries
 * This chapter describes the queries that EFML can perform against each device.
 * In each case the device is identified with an efmlDevice_t handle. This handle is obtained by
 * calling one of \ref efmlDeviceGetHandleByIndex_v2(), \ref efmlDeviceGetHandleBySerial(),
 * \ref efmlDeviceGetHandleByPciBusId_v2(). or \ref efmlDeviceGetHandleByUUID().
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Get the total number of devices.
 *
 * @param count [out] The total number of devices.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetCount_v2(unsigned int* count);

/**
 * @brief Get the device handle by index.
 *
 * @param index [in] The index of the device to get.
 * @param device [out] The device handle to return.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetHandleByIndex_v2(unsigned int index, efmlDevice_t* device);

/**
 * @brief Acquire the handle for a particular device, based on its UUID.
 *
 * @param uuid [in] The uuid of the device to get.
 * @param device [out] The device handle to return.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetHandleByUUID(const char* UUID, efmlDevice_t* device);

/**
 * @brief Acquire the handle for a particular device, based on its PCI bus id.
 *
 * @param pciBusId [in] The PCI bus id of the device to get.
 * @param device [out] The device handle to return.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetHandleByPciBusId_v2(const char* pciBusId, efmlDevice_t* device);

/**
 * @brief Retrieves the device's name
 *
 * @param device [in] The device handle.
 * @param name [out] The device's name.
 * @param length [in] The length of the name buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetName(efmlDevice_t device, char* name, unsigned int length);

/**
 * @brief Retrieves the index of the device.
 *
 * @param device [in] The device handle.
 * @param index [out] The index of the device.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetIndex(efmlDevice_t device, unsigned int *index);

/**
 * @brief Retrieves the device's serial number
 *
 * @param device [in] The device handle.
 * @param serial [out] The device's serial number.
 * @param length [in] The length of the serial number buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSerial(efmlDevice_t device, char *serial, unsigned int length);

/***************************************************************************************************/

/** @defgroup efmlAffinity CPU and Memory Affinity
 *  This chapter describes EFML operations that are associated with CPU and memory
 *  affinity.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Retrieves an array of unsigned ints (sized to cpuSetSize) of bitmasks with the ideal CPU affinity for the
 * device. For example, if processors 0, 1, 32, and 33 are ideal for the device and cpuSetSize == 2,
 *     result[0] = 0x3, result[1] = 0x3
 *
 * @param device [in] The device handle.
 * @param cpuSetSize [in] The size of the cpuSet array that is safe to access.
 * @param cpuSet [out] Array reference in which to return a bitmask of CPUs.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetCpuAffinity(efmlDevice_t device, unsigned int cpuSetSize,
    unsigned long *cpuSet);

/** @} */

/**
 * @brief Retrieves the UUID of the device.
 *
 * @param device [in] The device handle.
 * @param uuid [out] The device's UUID.
 * @param length [in] The length of the UUID buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetUUID(efmlDevice_t device, char* uuid, unsigned int length);

/**
 * @brief Retrieves minor number for the device.
 *
 * @param device [in] The device handle.
 * @param minorNumber [out] The minor number for the device.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMinorNumber(efmlDevice_t device, unsigned int *minorNumber);

/**
 * @brief Retrieves the board part number of the device.
 *
 * @param device [in] The device handle.
 * @param partNumber [out] The device's board part number.
 * @param length [in] The length of the board part number buffer.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetBoardPartNumber(efmlDevice_t device, char* partNumber, unsigned int length);

/**
 * @brief Retrieves the device's PCI information
 *
 * @param device [in] The device handle.
 * @param pciInfo [out] The device's PCI information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPciInfo_v3(efmlDevice_t device, efmlPciInfo_t* pciInfo);

/**
 * @brief Retrieves the maximum PCIe link generation possible with this device and system
 *
 * @param device [in] The device handle.
 * @param maxLinkGen [out] The maximum PCIe link generation.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMaxPcieLinkGeneration(efmlDevice_t device, unsigned int* maxLinkGen);

/**
 * @brief Retrieves the maximum PCIe link width possible with this device and system
 *
 * @param device [in] The device handle.
 * @param maxLinkWidth [out] The maximum PCIe link width.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMaxPcieLinkWidth(efmlDevice_t device, unsigned int* maxLinkWidth);

/**
 * @brief Retrieves the current PCIe link generation
 *
 * @param device [in] The device handle.
 * @param currLinkGen [out] The current PCIe link generation.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetCurrPcieLinkGeneration(efmlDevice_t device, unsigned int* currLinkGen);

/**
 * @brief Retrieves the current PCIe link width
 *
 * @param device [in] The device handle.
 * @param currLinkWidth [out] The current PCIe link width.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetCurrPcieLinkWidth(efmlDevice_t device, unsigned int* currLinkWidth);

/**
 * @brief Retrieve PCIe utilization information. This function is querying a byte counter over a 1ms interval and thus
 *        is the PCIe throughput over that interval.
 *
 * @param device [in] The device handle.
 * @param counter [in] The PCIe utilization counter to query.
 * @param value [out] Reference in which to return throughput in KB/s
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPcieThroughput(
    efmlDevice_t device, efmlPcieUtilCounter_t counter, unsigned int* value);

/**
 * @brief Retrieves the current clock speeds for the device.
 *
 * @param device [in] The device handle.
 * @param type [in] The type of clock to retrieve.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetClockInfo(efmlDevice_t device, efmlClockType_t type, unsigned int* clock);

/**
 * @brief Retrieves the maximum clock speeds for the device.
 *
 * @param device [in] The device handle.
 * @param type [in] The type of clock to retrieve.
 * @param clock [out] The maximum clock speed for the device in MHz.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMaxClockInfo(efmlDevice_t device, efmlClockType_t type, unsigned int* clock);

/**
 * @brief Retrieves the list of possible compute clocks that can be used as an argument for \ref efmlDeviceSetGcuLockedClocks.
 *
 * Supported only on Libra devices (GCU400 / GCU410 / GCU450). On other architectures returns EFML_ERROR_NOT_SUPPORTED_V2.
 *
 * @param device [in] The device handle.
 * @param memoryClockMHz [in] Not Used, just for compatibility.
 * @param count [in,out] On input, the number of \c unsigned int elements allocated in \p clocksMHz. On success, overwritten
 *                       with the number of frequencies written. If the buffer is too small, returns
 *                       EFML_ERROR_INSUFFICIENT_SIZE and sets \p *count to the minimum number of elements required.
 * @param clocksMHz [out] The list of possible compute clocks in MHz.
 *
 * @return EFML_SUCCESS_V2 on success, EFML_ERROR_INSUFFICIENT_SIZE if the array is shorter than required, or another error code on failure.
 *
 * @see efmlDeviceSetGcuLockedClocks
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSupportedComputeClocks(efmlDevice_t device, unsigned int memoryClockMHz, unsigned int *count, unsigned int *clocksMHz);

EFML_EXPORT efmlReturn_v1_t efmlDeviceGetFanSpeed(efmlDevice_t device, unsigned int* speed);
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetFanSpeed_v2(efmlDevice_t device, unsigned int fan, unsigned int* speed);

/**
 * @brief Retrieves the current temperature readings for the device, in degrees C.
 *
 * @param device [in] The device handle.
 * @param sensorType [in] The type of temperature sensor to retrieve.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetTemperature(efmlDevice_t device, efmlTemperatureSensors_t sensorType,
    int* temperature);

/**
 * @brief Retrieves the temperature threshold for the GCU with the specified threshold type in degrees C.
 *
 * @param device [in] The device handle.
 * @param thresholdType [in] The type of temperature threshold to retrieve.
 * @param temperature [out] The temperature threshold in degrees C.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetTemperatureThreshold(efmlDevice_t device,
    efmlTemperatureThresholds_t thresholdType, unsigned int* temperature);

/**
 * @brief Retrieves the current low power mode for the device.
 *
 * @param device [in] The device handle.
 * @param lpMode [out] Reference in which to return the low power mode
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetLowPowerMode(efmlDevice_t device, efmlLPModes_t *lpMode);

/**
 * @brief Retrieves the current low power mode enabled status for the device.
 *
 * @param device [in] The device handle.
 * @param enabled [out] Reference in which to return the low power mode enabled status
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetLowPowerModeEnabled(efmlDevice_t device, efmlEnableState_t *enabled);

/**
 * @brief Get all supported low power modes for the device.
 *
 * The returned array would contain a contiguous list of valid low power modes supported by the device.
 * If the number of supported low power modes is fewer than the size of the array supplied missing elements would contain EFML_LP_UNKNOWN.
 * The number of elements in the returned list will never exceed EFML_MAX_GCU_LOW_POWER_MODES.
 *
 * @param device [in] The device handle.
 * @param lpModes [out] Container to return the list of low power modes supported by device
 * @param size [in] Size of the supplied low power modes array
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSupportedLowPowerModes(efmlDevice_t device, efmlLPModes_t *lpModes,
    unsigned int size);

/**
 * @brief Retrieves the current power management limit for the device.
 *
 * @param device [in] The device handle.
 * @param limit [out] The current power management limit in milliwatts.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPowerManagementLimit(efmlDevice_t device, unsigned int* limit);


/**
 * @brief Retrieves the power management limit constraints for the device. Only supported on GCU400.
 *
 * @param device [in] The device handle.
 * @param minLimit [out] The minimum power management limit in milliwatts.
 * @param maxLimit [out] The maximum power management limit in milliwatts.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPowerManagementLimitConstraints(efmlDevice_t device, unsigned int* minLimit,
    unsigned int* maxLimit);

/**
 * @brief Retrieves the default power management limit for the device.
 *
 * @param device [in] The device handle.
 * @param defaultLimit [out] The default power management limit in milliwatts.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPowerManagementDefaultLimit(efmlDevice_t device, unsigned int* defaultLimit);

/**
 * @brief Retrieves power usage for this GCU in milliwatts
 *
 * @param device [in] The device handle.
 * @param power [out] The power usage in milliwatts
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetPowerUsage(efmlDevice_t device, unsigned int* power);

/**
 * @brief Get the effective power limit enforced by the driver.
 *
 * @param device [in] The device handle.
 * @param limit [out] The effective power limit in milliwatts.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetEnforcedPowerLimit(efmlDevice_t device, unsigned int* limit);

/**
 * @brief Retrieves the amount of used, free, reserved and total memory available on the device, in bytes.
 *        Enabling ECC reduces the amount of total available memory, due to the extra required parity bits.
 *
 * @param device [in] The device handle.
 * @param memory [out] The memory information for the device.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMemoryInfo(efmlDevice_t device, efmlMemory_t* memory);
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMemoryInfo_v2(efmlDevice_t device, efmlMemory_v2_t* memory);

/**
 * @brief Retrieves the Tops compute capability of the device.
 *
 * @param device [in] The device handle.
 * @param major [out] The major version of the compute capability.
 * @param minor [out] The minor version of the compute capability.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetTopsComputeCapability(efmlDevice_t device, int* major, int* minor);

/**
 * @brief Retrieves the ECC mode of the device.
 *
 * @param device [in] The device handle.
 * @param current [out] The current ECC mode.
 * @param pending [out] The pending ECC mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetEccMode(efmlDevice_t device, efmlEnableState_t *current,
    efmlEnableState_t *pending);

/**
 * @brief Retrieves the default ECC mode of the device.
 *
 * @param device [in] The device handle.
 * @param defaultMode [out] The default ECC mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetDefaultEccMode(efmlDevice_t device, efmlEnableState_t *defaultMode);

/**
 * @brief Retrieves the MCM mode of the device.
 *
 * @param device [in] The device handle.
 * @param mcm_mode [out] The MCM mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMcmMode(efmlDevice_t device, efmlMcmMode_t *mcm_mode);

/**
 * @brief Retrieves the memory error counter for the device.
 *
 * @param device [in] The device handle.
 * @param errorType [in] The type of memory error to retrieve.
 * @param counterType [in] The type of ECC counter to retrieve.
 * @param locationType [in] The type of memory location to retrieve.
 * @param count [out] The memory error counter.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMemoryErrorCounter(efmlDevice_t device, efmlMemoryErrorType_t errorType,
    efmlEccCounterType_t counterType, efmlMemoryLocation_t locationType, unsigned long long *count);

/**
 * @brief Retrieves the device reset count from RAS info (gcu_query_info_ras.rsv_out.v19.reset_count).
 *
 * @param device [in] The device handle.
 * @param resetCount [out] The device reset count.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetResetCount(efmlDevice_t device, unsigned int* resetCount);

/**
 * @brief Get RAS error records for a device (device-handle based API).
 *
 * @param device [in] The device handle.
 * @param count [in] Maximum number of records to return.
 * @param p_records [out] Buffer to receive error records.
 * @param total_count [out] Total number of records available.
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetErrorRecords(efmlDevice_t device, uint32_t count,
    efmlErrorRecord_t *p_records, uint32_t *total_count);

/**
 * @brief Retrieves the current utilization rates for the device's major subsystems.
 *        The utilization rates are returned as a percentage of time over the past sample period during which
 *        one or more kernels was executing on the GCU, and global (device) memory was being read or written.
 *
 * @param device [in] The device handle.
 * @param utilization [out] The device's utilization rates.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetUtilizationRates(efmlDevice_t device, efmlUtilization_t* utilization);

/**
 * @brief Retrieves the current occupancy for the device's major subsystems.
 *
 * @param device [in] The device handle.
 * @param occupancy [out] The device's occupancy rates.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetOccupancy(efmlDevice_t device, efmlOccupancy_t* occupancy);

/**
 * @brief Retrieves the current occupancy for the encoder.
 *
 * @param device [in] The device handle.
 * @param occupancy [out] The current occupancy for the encoder.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetEncoderOccupancy(efmlDevice_t device, unsigned int* occupancy);

/**
 * @brief Retrieves the current utilization and sampling size in microseconds for the encoder.
 *
 * @param device [in] The device handle.
 * @param utilization [out] The current utilization for the encoder.
 * @param samplingPeriodUs [out] The sampling period in microseconds for the utilization counter.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetEncoderUtilization(efmlDevice_t device, unsigned int* utilization,
    unsigned int* samplingPeriodUs);

/**
 * @brief Retrieves the current occupancy for the decoder.
 *
 * @param device [in] The device handle.
 * @param occupancy [out] The current occupancy for the decoder.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetDecoderOccupancy(efmlDevice_t device, unsigned int* occupancy);

/**
 * @brief Retrieves the current utilization and sampling size in microseconds for the decoder.
 *
 * @param device [in] The device handle.
 * @param utilization [out] The current utilization for the decoder.
 * @param samplingPeriodUs [out] The sampling period in microseconds for the utilization counter.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetDecoderUtilization(efmlDevice_t device, unsigned int* utilization,
    unsigned int* samplingPeriodUs);

/**
 * @brief Get VBIOS version of the device.
 *
 * @param device [in] The device handle.
 * @param version [out] The device's VBIOS version.
 * @param length [in] The length of the version buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetVbiosVersion(efmlDevice_t device, char* version, unsigned int length);

efmlReturn_v1_t efmlDeviceGetComputeRunningProcesses_v3(
    efmlDevice_t device, unsigned int* infoCount, efmlProcessInfo_v1_t* info);

EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGraphicsRunningProcesses_v3(
    efmlDevice_t device, unsigned int* infoCount, efmlProcessInfo_v1_t* info);

EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMPSComputeRunningProcesses_v3(
    efmlDevice_t device, unsigned int* infoCount, efmlProcessInfo_v1_t* info);

/**
 * @brief Get samples for the device.
 *
 * @param device [in] The device handle.
 * @param type [in] The type of sample to get.
 * @param lastSeenTimeStamp [in] The timestamp of the last sample to get.
 * @param sampleValType [out] The type of the sample value.
 * @param sampleCount [out] The number of samples returned.
 * @param samples [out] The samples returned.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSamples(efmlDevice_t device, efmlSamplingType_t type,
    unsigned long long lastSeenTimeStamp, efmlValueType_t *sampleValType, unsigned int *sampleCount,
    efmlSample_t *samples);

/**
 * @brief Get BAR1 memory information for the device.
 *
 * @param device [in] The device handle.
 * @param bar1Memory [out] The BAR1 memory information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetBAR1MemoryInfo(efmlDevice_t device, efmlBAR1Memory_t *bar1Memory);

/**
 * @brief Get BAR2 memory information for the device.
 *
 * @param device [in] The device handle.
 * @param bar2Memory [out] The BAR2 memory information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetBAR2MemoryInfo(efmlDevice_t device, efmlBAR2Memory_t *bar2Memory);

/**
 * @brief Get fabric information associated with the device.
 *
 * @param device [in] The device handle.
 * @param gcuFabricInfo [out] Information about GCU fabric state.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuFabricInfo(efmlDevice_t device, efmlGcuFabricInfo_t *gcuFabricInfo);

/**
 * @brief Get fabric information associated with the device (versioned).
 *
 * @param device [in] The device handle.
 * @param gcuFabricInfo [out] Information about GCU fabric state.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuFabricInfoV(efmlDevice_t device, efmlGcuFabricInfoV_t *gcuFabricInfo);

/**
 * @brief Retrieve AP firmware version.
 *
 * @param device [in] The device handle.
 * @param version [out] The retrieved AP firmware version
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetApFirmwareVersion(efmlDevice_t device, char *version);

/**
 * @brief Retrieve AP firmware mode.
 *
 * @param device [in] The device handle.
 * @param isEnabled [out] Pointer to specify if AP firmware is enabled
 * @param defaultMode [out] Pointer to specify if AP firmware is supported by default on \a device
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetApFirmwareMode(efmlDevice_t device, unsigned int *isEnabled,
    unsigned int *defaultMode);

/**
 * @brief Retrieve SP firmware version.
 *
 * @param device [in] The device handle.
 * @param version [out] The retrieved SP firmware version
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSpFirmwareVersion(efmlDevice_t device, char *version);

/**
 * @brief Retrieve SP firmware mode.
 *
 * @param device [in] The device handle.
 * @param isEnabled [out] Pointer to specify if SP firmware is enabled
 * @param defaultMode [out] Pointer to specify if SP firmware is supported by default on \a device
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSpFirmwareMode(efmlDevice_t device, unsigned int *isEnabled,
    unsigned int *defaultMode);

/**
 * @brief Retrieve SSM firmware version.
 *
 * @param device [in] The device handle.
 * @param version [out] The retrieved SSM firmware version
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSsmFirmwareVersion(efmlDevice_t device, char *version);

/**
 * @brief Retrieve SSM firmware mode.
 *
 * @param device [in] The device handle.
 * @param isEnabled [out] Pointer to specify if SSM firmware is enabled
 * @param defaultMode [out] Pointer to specify if SSM firmware is supported by default on \a device
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSsmFirmwareMode(efmlDevice_t device, unsigned int *isEnabled,
    unsigned int *defaultMode);

/**
 * @brief Retrieve VPU firmware version.
 *
 * @param device [in] The device handle.
 * @param version [out] The retrieved VPU firmware version
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetVpuFirmwareVersion(efmlDevice_t device, char *version);

/**
 * @brief Retrieve VPU firmware mode.
 *
 * @param device [in] The device handle.
 * @param isEnabled [out] Pointer to specify if VPU firmware is enabled
 * @param defaultMode [out] Pointer to specify if VPU firmware is supported by default on \a device
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetVpuFirmwareMode(efmlDevice_t device, unsigned int *isEnabled,
    unsigned int *defaultMode);

/**
 * @} // @defgroup efmlDeviceQueries Device Queries
 */

/** @addtogroup efmlDeviceQueries
 *  @{
 */

/**
 * @brief Get architecture for device
 *
 * @param device [in] The device handle.
 * @param arch [out] Reference where architecture is returned.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetArchitecture(efmlDevice_t device, efmlDeviceArchitecture_t* arch);

/**
 * @brief Retrieves the current utilization and process ID
 *
 * @param device [in] The device handle.
 * @param utilization [out] Pointer to caller-supplied buffer in which guest process utilization samples are returned
 * @param processSamplesCount [in/out] Pointer to caller-supplied array size, and returns number of processes running
 * @param lastSeenTimeStamp [in] Return only samples with timestamp greater than lastSeenTimeStamp.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetProcessUtilization(efmlDevice_t device,
    efmlProcessUtilizationSample_t* utilization, unsigned int* processSamplesCount,
    unsigned long long lastSeenTimeStamp);


/**
 * @brief Get slot OAM name for the device.
 *
 * @param device [in] The device handle.
 * @param name [out] The slot OAM name.
 * @param length [in] The length of the name buffer.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSlotOamName(efmlDevice_t device, char *name, unsigned int length);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlDeviceCommands Device Commands
 *  This chapter describes EFML operations that change the state of the device.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Set ECC mode for the device.
 *
 * @param device [in] The device handle.
 * @param mode [in] The ECC mode to set.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetEccMode(efmlDevice_t device, efmlEnableState_t mode);

/**
 * @brief Set MCM mode for the device.
 *
 * @param device [in] The device handle.
 * @param mode [in] The MCM mode to set.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetMcmMode(efmlDevice_t device, efmlMcmMode_t mode);

typedef enum efmlClockLimitId_enum {
    EFML_CLOCK_LIMIT_ID_RANGE_START = 0xffffff00,
    EFML_CLOCK_LIMIT_ID_TDP,
    EFML_CLOCK_LIMIT_ID_UNLIMITED
} efmlClockLimitId_t;

/**
 * @brief Set clocks that device will lock to.
 * Sets the clocks that the device will be running at to the value in the range of minGcuClockMHz to maxGcuClockMHz.
 *
 * This can be called with a pair of integer clock frequencies in MHz, or a pair of /ref efmlClockLimitId_t values.
 * See the table below for valid combinations of these values.
 *
 * minGcuClock | maxGcuClock | Effect
 * ------------+-------------+--------------------------------------------------
 *     tdp     |     tdp     | Not supported now
 *  unlimited  |     tdp     | Upper bound is TDP but clock may drift below this
 *     tdp     |  unlimited  | Not supported now
 *  unlimited  |  unlimited  | Unlocked (== efmlDeviceResetGcuLockedClocks)
 *
 * If one arg takes one of these values, the other must be one of these values as
 * well. Mixed numeric and symbolic calls return EFML_ERROR_INVALID_ARGUMENT_V2.
 *
 * @param device [in] The device handle.
 * @param minGcuClock [in] The minimum clock frequency in MHz. Only 25MHz is supported now.
 * @param maxGcuClock [in] The maximum clock frequency in MHz.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetGcuLockedClocks(efmlDevice_t device, unsigned int minGccuClockMHz, unsigned int maxGccuClockMHz);

/**
 * @brief Resets the gpu clock to the default value.
 *
 * @see efmlDeviceSetGcuLockedClocks
 *
 * @param device [in] The device handle.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceResetGcuLockedClocks(efmlDevice_t device);

/**
 * @brief Enable or disable low power mode for the device.
 *        If enabled, the device will automatically enter LP1/LP2 after being idle for a while. And can be set to
 *        LP3/LP0 manually.
 *        If disabled, the device will always stay in LP0.
 *
 * @param device [in] The device handle.
 * @param enable [in] The enable state of low power mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetLowPowerModeEnabled(efmlDevice_t device, efmlEnableState_t enable);

/**
 * @brief Set low power mode for the device
 *
 * @param device [in] The device handle.
 * @param mode [in] The low power mode to set.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetLowPowerMode(efmlDevice_t device, efmlLPModes_t mode);

/*
 * @brief Set power management limit for the device. Only supported on GCU400.
 *
 * @param device [in] The device handle.
 * @param limit [in] The power management limit.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetPowerManagementLimit(efmlDevice_t device, unsigned int limit);

/**
 * @}
 */

/***************************************************************************************************/
/** @defgroup GCU-LARE GCU-LARE Methods
 * This chapter describes methods that EFML can perform on GCU-LARE enabled devices.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Retrieves the state of the device's GCU-LARE for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param isActive [out] The state of the GCU-LARE for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareState(efmlDevice_t device, unsigned int port,
    efmlEnableState_t *isActive);

/**
 * @brief Retrieves the requested capability from the device's GcuLare for the port specified.
 *
 * @param device [in] The device handle.
 * @param port [in] The port number to query.
 * @param capability [in] The capability to query.
 * @param capResult [out] The result of the capability query.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareCapability(efmlDevice_t device, unsigned int port,
    efmlGcuLareCapability_t capability, unsigned int* capResult);

/**
 * @brief Retrieves the PCI information for the remote device connected over the given port.
 *
 * @param device [in] The device handle.
 * @param link [in] The link id to query.
 * @param pci [out] The PCI information for the remote device connected over the given port.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareRemotePciInfo_v2(efmlDevice_t device, unsigned int  link,
    efmlPciInfo_t* pci);

/**
 * @brief Get the GCU-LARE device type of the remote device connected over the given port.
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param deviceType [out] The GCU-LARE device type of the remote device connected over the given port.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareRemoteDeviceType(efmlDevice_t device, unsigned int port,
        efmlIntGcuLareDeviceType_t* deviceType);

efmlReturn_v1_t efmlDeviceTestGcuLareConnection(efmlDevice_t device,
    unsigned int source_address, unsigned int dest_address, int timeout_in_s, bool * result);

/**
 * @brief Retrieves the GCU-LARE information for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param info [out] The GCU-LARE information for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareInfo(efmlDevice_t device, unsigned int port, efmlGcuLareInfo_t *info);

/**
 * @brief Retrieves the remote port id for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param remotePort [out] The remote port id for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareRemotePort(efmlDevice_t device, unsigned int port,
    unsigned int* remotePort);


/**
 * @brief Retrieves the GCU-LARE QOS parameters for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param qosParam [out] The GCU-LARE QOS parameters for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareQosParam(efmlDevice_t device, unsigned int port,
    efmlGcuLareQosParam_t* qosParam);


/**
 * @brief Set the GCU-LARE QOS parameters for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to set the QOS parameters for.
 * @param buffer_id [in] The PFC buffer ID to set the QOS parameters for.
 * @param dscp [in] The DSCP value to set for the port.
 * @param priority [in] The priority value to set for the port.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetGcuLareQosParam(efmlDevice_t device, unsigned int port, unsigned int buffer_id,
    unsigned int dscp, unsigned int priority);

/**
 * @brief Set the GCU-LARE pause buffer threshold for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to set the pause buffer threshold for.
 * @param xon [in] The XON value to set for the port.
 * @param xoff [in] The XOFF value to set for the port.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetGcuLarePauseBufferThreshold(efmlDevice_t device, unsigned int port,
    unsigned int xon, unsigned int xoff);

/**
 * @brief Retrieves the transaction configuration for the GCU-LARE port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param config [out] The transaction configuration for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareTransactionConfig(efmlDevice_t device, unsigned int port,
    efmlGcuLareTransactionConfig_t* config);

/**
 * @brief Sets the timeout for the GCU-LARE port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to set the transaction configuration for.
 * @param timeout [in] The timeout to set for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetGcuLareTransactionTimeout(efmlDevice_t device, unsigned int port,
    unsigned long long timeout);

/**
 * @brief Sets the retry count for the GCU-LARE port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to set the retry count for.
 * @param retry_cnt [in] The retry count to set for the port specified.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetGcuLareTransactionRetryCount(efmlDevice_t device, unsigned int port,
    unsigned int retry_cnt);


/**
 * @brief Retrieves all the GCU-LARE queue IDs for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param queueCount [out] The number of GCU-LARE queues for the port specified.
 * @param queueID [out] The array of GCU-LARE queue IDs for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareQueues(efmlDevice_t device, unsigned int port, unsigned int* queueCount,
    unsigned int* queueID);

/**
 * @brief Retrieves the details of a specific GCU-LARE queue for the port specified
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to query.
 * @param queueId [in] The GCU-LARE queue ID to query.
 * @param queueDetail [out] The details of the GCU-LARE queue for the port specified.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuLareQueueDetails(efmlDevice_t device, unsigned int port,
    unsigned int queueId, efmlGcuLareQueueDetail_t* queueDetail);

/**
 * @brief Binds a GCU-LARE queue to a PFC buffer.
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to bind the queue to.
 * @param queueId [in] The GCU-LARE queue ID to bind.
 * @param pfcBufferId [in] The PFC buffer ID to bind the queue to.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceBindQueueToPfcBuffer(efmlDevice_t device, unsigned int port, unsigned int queueId,
    unsigned int pfcBufferId);

/**
 * @brief Isolate or connect a GCU-LARE port.
 *
 * @param device [in] The device handle.
 * @param port [in] The port id to isolate or connect.
 * @param isolate [in] The isolate value to set for the port. 1 to isolate, 0 to connect.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceIsolateGcuLarePort(efmlDevice_t device, unsigned int port, unsigned int isolate);

/** @} */ // @defgroup GCU-LARE GCU-LARE Methods

/***************************************************************************************************/
/** @defgroup efmlEvents Event Handling Methods
 * This chapter describes methods that EFML can perform against each device to register and wait for
 * some event to occur.
 *  @{
 */
/***************************************************************************************************/

/**
 * Create an empty set of events.
 * Event set should be freed by \ref efmlEventSetFree
 *
 * @param set [out] Reference in which to return the event handle
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @see efmlEventSetFree
 */
EFML_EXPORT efmlReturn_v1_t efmlEventSetCreate(efmlEventSet_t *set);

/**
 * Starts recording of events on a specified devices and add the events to specified \ref efmlEventSet_t
 *
 * This call starts recording of events on specific device.
 * All events that occurred before this call are not recorded.
 * Checking if some event occurred can be done with \ref efmlEventSetWait_v2
 *
 * If function reports EFML_ERROR_UNKNOWN, event set is in undefined state and should be freed.
 * If function reports EFML_ERROR_NOT_SUPPORTED_V2, event set can still be used. None of the requested eventTypes
 *     are registered in that case.
 *
 * @param device [in] The identifier of the target device
 * @param eventTypes [in] Bitmask of \ref efmlEventType to record
 * @param set [in] Set to which add new event types
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @see efmlEventType
 * @see efmlDeviceGetSupportedEventTypes
 * @see efmlEventSetWait
 * @see efmlEventSetFree
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceRegisterEvents(efmlDevice_t device, unsigned long long eventTypes, efmlEventSet_t set);

/**
 * Returns information about events supported on device
 *
 * @param device [in] The identifier of the target device
 * @param eventTypes [out] Reference in which to return bitmask of supported events
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @see efmlEventType
 * @see efmlDeviceRegisterEvents
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetSupportedEventTypes(efmlDevice_t device, unsigned long long *eventTypes);

/**
 * Waits on events and delivers events
 *
 * If some events are ready to be delivered at the time of the call, function returns immediately.
 * If there are no events ready to be delivered, function sleeps till event arrives
 * but not longer than specified timeout. This function in certain conditions can return before
 * specified timeout passes (e.g. when interrupt arrives)
 *
 * @param set [in] Reference to set of events to wait on
 * @param data [out] Reference in which to return event data
 * @param timeoutms [in] Maximum amount of wait time in milliseconds for registered event
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @see efmlEventType
 * @see efmlDeviceRegisterEvents
 */
EFML_EXPORT efmlReturn_v1_t efmlEventSetWait_v2(efmlEventSet_t set, efmlEventData_t * data, unsigned int timeoutms);

/**
 * Releases events in the set
 *
 * @param set [in] Reference to events to be released
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 *
 * @see efmlDeviceRegisterEvents
 */
EFML_EXPORT efmlReturn_v1_t efmlEventSetFree(efmlEventSet_t set);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlFieldValueQueries Field Value Queries
 *  This chapter describes EFML operations that are associated with retrieving Field Values from EFML
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Request values for a list of fields for a device. This API allows multiple fields to be queried at once.
 *
 * @param device [in] The device handle.
 * @param valuesCount [in] The number of fields to query.
 * @param values [out] The array of field values to return. Each value's fieldId must be populated prior to this call.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetFieldValues(efmlDevice_t device, int valuesCount, efmlFieldValue_t* values);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlVirtualGcuQueries vGCU APIs
 * This chapter describes operations that are associated with Enflame vGCU Software products.
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Retrieves the device's virtualization mode
 *
 * @param device [in] The device handle.
 * @param mode [out] The device's virtualization mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetVirtualizationMode(efmlDevice_t device, efmlGcuVirtualizationMode_t* mode);

/**
 * Queries if SR-IOV host operation is supported on a device.
 *
 * Checks whether SR-IOV host capability is supported by the device and the
 * driver, and indicates device is in SR-IOV mode if both of these conditions
 * are true.
 *
 * @param device [in] The identifier of the target device
 * @param mode [out] Reference in which to return the current GCU SR-IOV mode
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetHostVgcuMode(efmlDevice_t device, efmlHostVgcuMode_t *mode);

/**
 * @brief Retrieves the number of active vGCUs on the device
 *
 * @param device [in] The device handle.
 * @param count [out] Reference in which to return the number of active vGCUs.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetActiveVgcusCount(efmlDevice_t device, unsigned int* count);

/** @} */

/***************************************************************************************************/
/** @defgroup efmlDeviceResourceSlice Device Resource Slice Management
 * This chapter describes EFML operations that are associated with Device Resource Slice management.
 *  @{
 */
/***************************************************************************************************/

#define EFML_GCU_INSTANCE_PROFILE_1_SLICE      0x0
#define EFML_GCU_INSTANCE_PROFILE_2_SLICE      0x1
#define EFML_GCU_INSTANCE_PROFILE_3_SLICE      0x2
#define EFML_GCU_INSTANCE_PROFILE_4_SLICE      0x3
#define EFML_GCU_INSTANCE_PROFILE_6_SLICE      0x4
#define EFML_GCU_INSTANCE_PROFILE_12_SLICE     0x5
#define EFML_GCU_INSTANCE_PROFILE_24_SLICE     0x6
#define EFML_GCU_INSTANCE_PROFILE_COUNT        0x7

typedef struct efmlGcuInstancePlacement_st
{
    unsigned int start;               //!< Index of first occupied memory slice
    unsigned int size;                //!< Number of memory slices occupied
} efmlGcuInstancePlacement_t;

/**
 * GCU instance profile information.
 */
typedef struct efmlGcuInstanceProfileInfo_st
{
    unsigned int id;                        //!< Unique profile ID within the device
    unsigned int isP2pSupported;            //!< Peer-to-Peer support
    unsigned int sliceCount;                //!< GCU Slice count
    unsigned int instanceCount;             //!< GCU instance count
    unsigned int sipCount;                  //!< Scalable intelligent processor count
    unsigned int transformationEngineCount; //!< Transformation Engine count
    unsigned int decoderCount;              //!< Decoder Engine count
    unsigned int encoderCount;              //!< Encoder Engine count
    unsigned long long memorySizeMB;        //!< Memory size in MBytes
} efmlGcuInstanceProfileInfo_t;

/**
 * GCU instance profile information (v2).
 *
 * Version 2 adds the \ref efmlGcuInstanceProfileInfo_v2_t.version field
 * to the start of the structure, and the \ref efmlGcuInstanceProfileInfo_v2_t.name
 * field to the end. This structure is not backwards-compatible with
 * \ref efmlGcuInstanceProfileInfo_t.
 */
typedef struct efmlGcuInstanceProfileInfo_v2_st
{
    unsigned int version;                       //!< Structure version identifier (set to \ref efmlGcuInstanceProfileInfo_v2)
    unsigned int id;                            //!< Unique profile ID within the device
    unsigned int isP2pSupported;                //!< Peer-to-Peer support
    unsigned int sliceCount;                    //!< GCU Slice count
    unsigned int instanceCount;                 //!< GCU instance count
    unsigned int sipCount;                      //!< Scalable intelligent processor count
    unsigned int transformationEngineCount;     //!< Transformation Engine count
    unsigned int decoderCount;                  //!< Decoder Engine count
    unsigned int encoderCount;                  //!< Encoder Engine count
    unsigned long long memorySizeMB;            //!< Memory size in MBytes
    char name[EFML_DEVICE_NAME_V2_BUFFER_SIZE]; //!< Profile name
} efmlGcuInstanceProfileInfo_v2_t;

/**
 * Version identifier value for \ref efmlGcuInstanceProfileInfo_v2_t.version.
 */
#define efmlGcuInstanceProfileInfo_v2 EFML_STRUCT_VERSION(GcuInstanceProfileInfo, 2)


typedef struct efmlGcuInstanceInfo_st
{
    efmlDevice_t device;                      //!< Parent device
    unsigned int id;                          //!< Unique instance ID within the device
    unsigned int profileId;                   //!< Unique profile ID within the device
    efmlGcuInstancePlacement_t placement;     //!< Placement for this instance
    uint8_t uuid[16];                         //!< UUID of the GCU instance
    uint64_t totalMemorySize;
    uint64_t usedMemorySize;
} efmlGcuInstanceInfo_t;

typedef struct efmlGcuInstance_st* efmlGcuInstance_t;

/**
 * @brief Wrapper for efmlDeviceSetDrsMode
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetMigMode(efmlDevice_t device, unsigned int mode,
    efmlReturn_v1_t* activationStatus);

/**
 * @brief Set DRS mode for the device.
 *
 * @param device [in] The device handle.
 * @param mode [in] The DRS mode to set.
 * @param activationStatus [out] The activation status of the DRS mode. If device is not idle, EFML_ERROR_IN_USE would
 *        be returned.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceSetDrsMode(efmlDevice_t device, unsigned int mode,
    efmlReturn_v1_t* activationStatus);

/**
 * @brief Wrapper for efmlDeviceGetDrsMode
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetMigMode(efmlDevice_t device, unsigned int* currentMode,
    unsigned int* pendingMode);

/**
 * @brief Get DRS mode for the device.
 *
 * @param device [in] The device handle.
 * @param currentMode [out] The current DRS mode.
 * @param pendingMode [out] The pending DRS mode.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetDrsMode(efmlDevice_t device, unsigned int* currentMode,
    unsigned int* pendingMode);

/**
 * @brief Get the GCU instance profile information for the device.
 *
 * @param device [in] The device handle.
 * @param profile [in] The GCU instance profile ID. One of the EFML_GCU_INSTANCE_PROFILE_*
 * @param info [out] The GCU instance profile information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuInstanceProfileInfo(efmlDevice_t device, unsigned int profile,
    efmlGcuInstanceProfileInfo_t* info);

/**
 * @brief Versioned wrapper around efmlDeviceGetGcuInstanceProfileInfo that accepts a versioned efmlGcuInstanceProfileInfo_v2_t or later output structure.
 *
 * @param device [in] The device handle.
 * @param profile [in] The GCU instance profile ID. One of the EFML_GCU_INSTANCE_PROFILE_*
 * @param info [out] The GCU instance profile information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuInstanceProfileInfoV(efmlDevice_t device, unsigned int profile,
    efmlGcuInstanceProfileInfo_v2_t* info);

/**
 * @brief Create a GCU instance for the device.
 *
 * @param device [in] The device handle.
 * @param profileId [in] The GCU instance profile ID. See efmlDeviceGetGcuInstanceProfileInfo
 * @param gcuInstance [out] The GCU instance handle.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceCreateGcuInstance(efmlDevice_t device, unsigned int profileId,
    efmlGcuInstance_t* gcuInstance);

/**
 * @brief Destroy GCU instance.
 *
 * @param gcuInstance [in] The GCU instance handle.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGcuInstanceDestroy(efmlGcuInstance_t gcuInstance);

/**
 * @brief Get GCU instances for given profile ID.
 *
 * @param device [in] The device handle.
 * @param profileId [in] The GCU instance profile ID. See efmlDeviceGetGcuInstanceProfileInfo
 * @param gcuInstances [out] The GCU instance handles.
 * @param count [out] The number of GCU instances returned.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuInstances(efmlDevice_t device, unsigned int profileId,
    efmlGcuInstance_t* gcuInstances, unsigned int* count);

/**
 * @brief Get GCU instance by ID.
 *
 * @param device [in] The device handle.
 * @param id [in] The GCU instance ID.
 * @param gcuInstance [out] The GCU instance handle.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlDeviceGetGcuInstanceById(efmlDevice_t device, unsigned int id,
    efmlGcuInstance_t* gcuInstance);

/**
 * @brief Get GCU instance information.
 *
 * @param gcuInstance [in] The GCU instance handle.
 * @param info [out] The GCU instance information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGcuInstanceGetInfo(efmlGcuInstance_t gcuInstance, efmlGcuInstanceInfo_t* info);

/** @} */  // @defgroup efmlDeviceResourceSlice

/***************************************************************************************************/
/** @defgroup GPM EFML GPM
 *  @{
 */
/***************************************************************************************************/
/** @defgroup efmlGpmEnums GPM Enums
 *  @{
 */
/***************************************************************************************************/

/**
 * GPM Metric Identifiers
 */

typedef enum
{
    EFML_GPM_METRIC_SIP_UTIL               = 2,    //!< Percentage of SIPs that were busy. 0.0 - 100.0
    EFML_GPM_METRIC_SIP_OCCUPANCY          = 3,    //!< Percentage of threads that were active vs theoretical maximum. 0.0 - 100.0
    EFML_GPM_METRIC_SPECIAL_FUNCTION_UTIL  = 6,    //!< Percentage of time the GCU's SIPs were doing special function operations. 0.0 - 100.0
    EFML_GPM_METRIC_MATRIX_UTIL            = 7,    //!< Percentage of time the GCU's SIPs were doing 2D operations. 0.0 - 100.0
    EFML_GPM_METRIC_VECTOR_UTIL            = 9,    //!< Percentage of time the GCU's SIPs were doing 1D operations. 0.0 - 100.0
    EFML_GPM_METRIC_DRAM_BW_UTIL           = 10,   //!< Percentage of DRAM bw used vs theoretical maximum. 0.0 - 100.0 */
    EFML_GPM_METRIC_DRAM_READ_PER_SEC      = 18,   //!< DRAM read traffic from this GCU in MiB/sec
    EFML_GPM_METRIC_DRAM_WRITE_PER_SEC     = 19,   //!< DRAM write traffic to this GCU in MiB/sec
    EFML_GPM_METRIC_PCIE_TX_PER_SEC        = 20,   //!< PCIe traffic from this GCU in MiB/sec
    EFML_GPM_METRIC_PCIE_RX_PER_SEC        = 21,   //!< PCIe traffic to this GCU in MiB/sec
    EFML_GPM_METRIC_TOPSDEC_0_UTIL         = 30,   //!< Percent utilization of TOPSDEC 0. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_1_UTIL         = 31,   //!< Percent utilization of TOPSDEC 1. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_2_UTIL         = 32,   //!< Percent utilization of TOPSDEC 2. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_3_UTIL         = 33,   //!< Percent utilization of TOPSDEC 3. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_4_UTIL         = 34,   //!< Percent utilization of TOPSDEC 4. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_5_UTIL         = 35,   //!< Percent utilization of TOPSDEC 5. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_6_UTIL         = 36,   //!< Percent utilization of TOPSDEC 6. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSDEC_7_UTIL         = 37,   //!< Percent utilization of TOPSDEC 7. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSENC_0_UTIL         = 166,  //!< Percent utilization of TOPSENC 0. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSENC_1_UTIL         = 167,  //!< Percent utilization of TOPSENC 1. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSENC_2_UTIL         = 168,  //!< Percent utilization of TOPSENC 2. 0.0 - 100.0
    EFML_GPM_METRIC_TOPSENC_3_UTIL         = 169,  //!< Percent utilization of TOPSENC 3. 0.0 - 100.0
    EFML_GPM_METRIC_MAX                            //!< Maximum value above +1.
} efmlGpmMetricId_t;

/** @} */ // @defgroup efmlGpmEnums

/***************************************************************************************************/
/** @defgroup efmlGpmStructs GPM Structs
 *  @{
 */
/***************************************************************************************************/

/**
 * Handle to an allocated GPM sample allocated with efmlGpmSampleAlloc(). Free this with efmlGpmSampleFree().
 */
typedef struct efmlGpmSample_st* efmlGpmSample_t;

/**
 * GPM metric information.
 */
typedef struct
{
    unsigned int metricId;      //!<  IN: EFML_GPM_METRIC_? define of which metric to retrieve
    efmlReturn_v1_t efmlReturn; //!<  OUT: Status of this metric. If this is nonzero, then value is not valid
    double value;               //!<  OUT: Value of this metric. Is only valid if efmlReturn is 0 (EFML_SUCCESS_V2)
    struct
    {
        char *shortName;
        char *longName;
        char *unit;
    } metricInfo;               //!< OUT: Metric name and unit. Those can be NULL if not defined
} efmlGpmMetric_t;

/**
 * GPM buffer information.
 */
typedef struct
{
    unsigned int version;                              //!< IN: Set to EFML_GPM_METRICS_GET_VERSION
    unsigned int numMetrics;                           //!< IN: How many metrics to retrieve in metrics[]
    efmlGpmSample_t sample1;                           //!< IN: Sample buffer
    efmlGpmSample_t sample2;                           //!< IN: Sample buffer
    efmlGpmMetric_t metrics[EFML_GPM_METRIC_MAX];      //!< IN/OUT: Array of metrics. Set metricId on call. See efmlReturn and value on return
} efmlGpmMetricsGet_t;

#define EFML_GPM_METRICS_GET_VERSION 1

/**
 * GPM device information.
 */
typedef struct
{
    unsigned int version;           //!< IN: Set to EFML_GPM_SUPPORT_VERSION
    unsigned int isSupportedDevice; //!< OUT: Indicates device support
} efmlGpmSupport_t;

#define EFML_GPM_SUPPORT_VERSION 1

/** @} */ // @defgroup efmlGpmStructs

/***************************************************************************************************/
/** @defgroup efmlGpmFunctions GPM Functions
 *  @{
 */
/***************************************************************************************************/

/**
 * @brief Calculate GPM metrics from two samples.
 *
 * To retrieve metrics, the user must first allocate the two sample buffers at \a metricsGet->sample1
 * and \a metricsGet->sample2 by calling \a efmlGpmSampleAlloc(). Next, the user should fill in the ID of each metric
 * in \a metricsGet->metrics[i].metricId and specify the total number of metrics to retrieve in \a metricsGet->numMetrics,
 * The version should be set to EFML_GPM_METRICS_GET_VERSION in \a metricsGet->version. The user then calls the
 * \a efmlGpmSampleGet() API twice to obtain 2 samples of counters. \note that the interval between these
 * two \a efmlGpmSampleGet() calls should be greater than 50ms due to the internal sample refresh rate.
 * Finally, the user calls \a efmlGpmMetricsGet to retrieve the metrics, which will be stored at \a metricsGet->metrics
 *
 * @param metricsGet [in/out] The GPM metrics to get.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGpmMetricsGet(efmlGpmMetricsGet_t *metricsGet);

/**
 * @brief Free an allocated sample buffer that was allocated with \ref efmlGpmSampleAlloc()
 *
 * @param gpmSample [in] The GPM sample to free.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGpmSampleFree(efmlGpmSample_t gpmSample);

/**
 * @brief Allocate a sample buffer to be used with EFML GPM . You will need to allocate
 * at least two of these buffers to use with the EFML GPM feature
 *
 * @param gpmSample [out] The GPM sample to allocate.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGpmSampleAlloc(efmlGpmSample_t *gpmSample);

/**
 * @brief Read a sample of GPM metrics into the provided \a gpmSample buffer. After
 * two samples are gathered, you can call efmlGpmMetricGet on those samples to
 * retrive metrics
 *
 * @note The interval between two \a efmlGpmSampleGet() calls should be greater than 50ms due to
 * the internal sample refresh rate.
 *
 * @param device [in] The device to get the GPM sample from.
 * @param gpmSample [out] The GPM sample to get.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGpmSampleGet(efmlDevice_t device, efmlGpmSample_t gpmSample);

/**
 * @brief Indicate whether the supplied device supports GPM
 *
 * @param device [in] The device to query.
 * @param gpmSupport [out] The GPM support information.
 *
 * @return EFML_SUCCESS_V2 on success, or an error code on failure.
 */
EFML_EXPORT efmlReturn_v1_t efmlGpmQueryDeviceSupport(efmlDevice_t device, efmlGpmSupport_t *gpmSupport);

/** @} */ // @defgroup efmlGpmFunctions
/** @} */ // @defgroup GPM

#ifndef DOXYGEN_SHOULD_SKIP_THIS
efmlReturn_v1_t efmlDeviceGetChipVersion(efmlDevice_t device, unsigned int* chipVersion);
#endif /* DOXYGEN_SHOULD_SKIP_THIS */

#ifdef __cplusplus
}  // extern "C"
#endif

#include "efml/efml_legacy.h"

#endif  // LIB_EFML_INCLUDE_EFML_H_
