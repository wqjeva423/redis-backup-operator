import os
import sys
import json
import subprocess
from datetime import datetime, timezone, timedelta
from redis.sentinel import Sentinel

master_name = "mymaster"
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
backup_start_time = datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y-%m-%d %H:%M:%S")
backup_filename = "redis-bak-" + cluster_name + "-" + datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y%m%d%H%M%S") + ".tar.gz"
redis_host = os.getenv('RFS_%s_SERVICE_HOST' % str.upper(cluster_name))
redis_port = os.getenv('RFS_%s_SERVICE_PORT' % str.upper(cluster_name))
file_path = S3BackupURL+backup_filename


def backup():
    os.chdir("/data")
    status_tar, res_tar = subprocess.getstatusoutput("tar -zcvf /tmp/" + backup_filename + " *.*")
    if status_tar != 0:
        print("backup tar error %s: %s" % (status_tar, res_tar))
    else:
        print("backup tar succeed: %s" % res_tar)
    status_file, filesize = subprocess.getstatusoutput("du -sh /tmp/" + backup_filename + " |awk '{print $1}'")
    status_s3, res_s3 = subprocess.getstatusoutput("s3cmd --no-ssl --region=%s --host=%s "
                                                   "--host-bucket=%s --access_key=%s "
                                                   "--secret_key=%s put /tmp/redis-bak-* %s" %
                                                   (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey, S3BackupURL))

    backup_end_time = datetime.now().astimezone(timezone(timedelta(hours=8))).strftime("%Y-%m-%d %H:%M:%S")
    if status_s3 != 0:
        print("backup S3 put error %s: %s" % (status_s3, res_s3))
        backup_data = [{
            "filePath": file_path,
            "fileSize": filesize,
            "jobName": jobname,
            "podName": hostname,
            "backupStartTime": backup_start_time,
            "backupEndTime": backup_end_time,
            "objectName": object_name,
            "status": "Failed",
            "message": res_s3
        }]
        return json.dumps(backup_data)
    else:
        print("backup S3 put succeed: %s" % res_s3)
    status_rm, res_rm = subprocess.getstatusoutput("rm -rf /tmp/redis-bak-*")

    backup_data = [{
        "filePath": file_path,
        "fileSize": filesize,
        "jobName": jobname,
        "podName": hostname,
        "backupStartTime": backup_start_time,
        "backupEndTime": backup_end_time,
        "objectName": object_name,
        "status": "Succeed",
        "message": "None"
    }]
    return json.dumps(backup_data)


def redis_conn(keyname, redis_data):
    try:
        sentinel = Sentinel([(redis_host, redis_port)], socket_timeout=0.5)
        master = sentinel.master_for(master_name, socket_timeout=0.5, password=password, db=15)
        master.set(keyname, redis_data)
        if keyname != "backupCronjob":
            master.expire(keyname, 3600)
        print("set redis data: %s " % redis_data)
    except Exception as e:
        print("redis conn error: %s" % str(e))


if __name__ == '__main__':
    data = backup()
    redis_conn(sys.argv[1], data)
