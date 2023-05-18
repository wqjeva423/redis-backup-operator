import os
import sys
import json
import subprocess
from datetime import datetime, timezone, timedelta
from rediscluster import RedisCluster

primaries_number = os.getenv("numberOfPrimaries")
password = os.getenv("redisPass")
hostname = os.getenv("HOSTNAME")
jobname = os.getenv("jobName")
object_name = os.getenv("objectName")
cluster_name = os.getenv("redisClusterName")
S3Region = os.getenv("S3Region")
S3Endpoint = os.getenv("S3Endpoint")
S3AccessKey = os.getenv("S3AccessKey")
S3SecretKey = os.getenv("S3SecretKey")
S3BackupURL = os.getenv("backupURL")
backup_redis_host = os.getenv("backupRedisHost")
backup_start_time = datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y-%m-%d %H:%M:%S")
backup_filename = "redis-bak-" + cluster_name + "-" + datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y%m%d%H%M%S") + "-N" + primaries_number


def backup(keyname):
    try:
        file_size = 0
        backup_file_list = []
        backup_data = {
            "filePath": backup_file_list,
            "fileSize": file_size,
            "jobName": jobname,
            "podName": hostname,
            "backupStartTime": backup_start_time,
            "backupEndTime": "",
            "objectName": object_name,
            "status": "Failed",
            "message": ""
        }

        status_ls, res_ls = subprocess.getstatusoutput("ls / |grep data")
        if len(res_ls.split('\n')) != int(primaries_number):
            backup_data['message'] = "Mount PVCs number not match Primaries number %s: %s" % (len(res_ls.split('\n')), primaries_number)
            redis_conn(keyname, json.dumps([backup_data]))
            raise Exception("Mount PVCs number not match Primaries number %s: %s" % (len(res_ls.split('\n')), primaries_number))
        else:
            print("------ Start Backup All PVCs Data ------")
            for i in res_ls.split('\n'):
                os.chdir("/"+i)
                status_tar, res_tar = subprocess.getstatusoutput("tar -zcvf " + backup_filename + "-" + i + ".tar.gz *")
                if status_tar != 0:
                    backup_data['message'] = "backup %s error %s: %s" % (i, status_tar, res_tar)
                    redis_conn(keyname, json.dumps([backup_data]))
                    raise Exception("backup %s error %s: %s" % (i, status_tar, res_tar))
                else:
                    print("backup %s succeed" % i)
                    backup_file_list.append(S3BackupURL + backup_filename + "-" + i + ".tar.gz")
                    status_file, filesize = subprocess.getstatusoutput("du -sm " + backup_filename + "-" + i + ".tar.gz |awk '{print $1}'")
                    file_size = file_size + int(filesize)

            print("------ Start Upload Backup Files to S3 ------")
            for i in res_ls.split('\n'):
                os.chdir("/" + i)
                status_s3, res_s3 = subprocess.getstatusoutput("s3cmd --no-ssl --region=%s --host=%s "
                                                           "--host-bucket=%s --access_key=%s "
                                                           "--secret_key=%s put redis-bak-* %s" %
                                                           (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey, S3BackupURL))
                if status_s3 != 0:
                    backup_data['message'] = "Upload Backups %s to S3 error %s: %s" % (i, status_s3, res_s3)
                    for del_bak in backup_file_list:
                        subprocess.getstatusoutput("s3cmd --no-ssl --region=%s --host=%s "
                                                   "--host-bucket=%s --access_key=%s "
                                                   "--secret_key=%s del %s" %
                                                   (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey,
                                                    del_bak))
                    redis_conn(keyname, json.dumps([backup_data]))
                    raise Exception("Upload Backups %s to S3 error %s: %s" % (i, status_s3, res_s3))
                else:
                    print("Upload Backups %s to S3 succeed" % i)
                    status_rm, res_rm = subprocess.getstatusoutput("rm -rf redis-bak-*")
            backup_end_time = datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y-%m-%d %H:%M:%S")
            backup_data['status'] = "Succeed"
            backup_data['backupEndTime'] = backup_end_time
            backup_data['fileSize'] = file_size
            redis_conn(keyname, json.dumps([backup_data]))
    except Exception as e:
        raise Exception("Backup main error: %s" % str(e))


def redis_conn(keyname, redis_data):
    try:
        print("------ Set Redis Data ------")
        startup_nodes = []
        for i in backup_redis_host.split(','):
            startup_nodes.append({'host': i, 'port': 6379})
        rc = RedisCluster(startup_nodes=startup_nodes, decode_responses=True, password=password, skip_full_coverage_check=True)
        if keyname != "backupCronjob":
            rc.set(keyname, redis_data, ex=3600)
        else:
            rc.set(keyname, redis_data)
        print("set redis data: %s " % redis_data)
    except Exception as e:
        print("redis conn error: %s" % str(e))


if __name__ == '__main__':
    data = backup(sys.argv[1])