import os
import subprocess
import sys

S3Region = os.getenv("S3Region")
S3Endpoint = os.getenv("S3Endpoint")
S3AccessKey = os.getenv("S3AccessKey")
S3SecretKey = os.getenv("S3SecretKey")
S3BackupURL = os.getenv("backupURL")
file_path = os.getenv("FILE_PATH")


def redis_recover():
    try:
        mk_status, mk_res = subprocess.getstatusoutput("mkdir -p /data")
        if mk_status != 0:
            raise Exception("mkdir /data failed %s: %s" % (mk_status, mk_res))
        os.chdir("/data")
        check_status, check_res = subprocess.getstatusoutput("ls /data")
        if check_status != 0:
            raise Exception("check data dir failed %s: %s" % (check_status, check_status))
        else:
            if len(check_res) > 0:
                print("data dir not empty. nothing to do.")
                return
            else:
                status_s3, res_s3 = subprocess.getstatusoutput("s3cmd --no-ssl --region=%s --host=%s "
                                                               "--host-bucket=%s --access_key=%s "
                                                               "--secret_key=%s get %s redis-bak.tar.gz" %
                                                               (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey,
                                                                file_path))
                if status_s3 != 0:
                    raise Exception("S3cmd get backup file error %s: %s" % (status_s3, res_s3))
                else:
                    print("S3cmd get backup file succeed: %s" % res_s3)
                    tar_status, tar_res = subprocess.getstatusoutput("tar -zxvf redis-bak.tar.gz")
                    if tar_status != 0:
                        raise Exception("tar backup file failed %s: %s" % (tar_status, tar_res))
                    else:
                        print("redis recover succeed")
                status_rm, res_rm = subprocess.getstatusoutput("rm -rf /data/redis-bak.tar.gz")
                return
    except Exception as e:
        print("redis recover error: " + str(e))
        sys.exit(-1)


if __name__ == '__main__':
    redis_recover()
