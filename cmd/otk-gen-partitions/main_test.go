package main_test

// idea 1 - full flexible, allows users to go wild (but ugly)
var fakeGenPartitionsInput1 = `
{
  "partition-table": {
    "uuid": "D209C89E-EA5E-4FBD-B161-B461CCE297E0",
    "type": "gpt",
    "partitions": [
      {
        "size": 1048576,
        "bootable": true,
        "type": "BIOSBootPartitionGUID",
        "uuid": "BIOSBootPartitionUUID"
      }, {
        "size": 209715200,
        "type": "EFISystemPartitionGUID",
        "uuid": "EFISystemPartitionUUID"
        "payload": {
          "type": "vfat",
          "uuid": "EFIFilesystemUUID",
          "mount_point": "/boot/efi",
          "label": "EFI-SYSTEM",
          "fstab": {
            "options": "defaults,uid=0,gid=0,umask=077,shortname=winnt",
            "freq": 0,
            "passno": 2
          }
        }
      }
    }
  }
}
`

// idea 2 - least flexible, quickest way to success
var fakeGenPartitionsInput2 = `
{
  "from-template": "centos-9-x86_64"
  "modifications": {
    "filesystem": [
      {
         "mountpoint": "/",
         "size": "20Gib"
      }
    ]
  }
}
`

// idea 3 - mix between (1) and (2), allow the details of (1) but have
// sensible defaults/shortcuts, works well for amd64/arm64 but less well
// for ppc64el/s390x which will have a very manual partition table
// (c.f. images/pkg/distro/rhel/rhel9/partiton_tables.go)
//
// TODO: how to pass "bootsize" in a nice way so that it reads nicely in
//
//	an otk file? bootsize is changed in 9.2, 9.4, 9.6 so making it
//	trivial in the otk to override easiyl without too much repetition
//	is important
var fakeGenPartitionsInput3_x86 = `
{
  "partition-table": {
    "type": "gpt",        
    "default_fs": "ext4"  
    "partitions": [       
      { "size": "1Mib", "type": "biosboot", "bootable": true },
      { "mountpoint": "/" },
      { "mountpoint": "/boot" },
      { "mountpoint": "/boot/efi" }
    ]
  ]
}`

var fakeGenPartitionsInput3_aarch64 = `
{
  "partition-table": {
    "type": "gpt",        
    "default_fs": "ext4"  
    "partitions": [       
      { "mountpoint": "/" },
      { "mountpoint": "/boot" },
      { "mountpoint": "/boot/efi" }
    ]
  ]
}`

var fakeGenPartitionsInput3_s390x = `
{
  "partition-table": {
    "type": "dos",
    "uuid": "0x14fc63d2",
    "default_fs": "xfs"
    "partitions": [
      { "size": "4Mib", "type": 41, "bootable": true },
      { "mountpoint": "/" },
      { "mountpoint": "/boot" },
    ]
  ]
}

`
