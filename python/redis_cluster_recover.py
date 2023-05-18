import os
import subprocess
import sys

S3Region = os.getenv("S3Region")
S3Endpoint = os.getenv("S3Endpoint")
S3AccessKey = os.getenv("S3AccessKey")
S3SecretKey = os.getenv("S3SecretKey")
S3RecoverURLs = os.getenv("S3RecoverURLs")


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
            s3_recover_urls = S3RecoverURLs.split(',')
            for s3_recover_url in s3_recover_urls:
                print("s3cmd --no-ssl --region=%s --host=%s "
                        "--host-bucket=%s --access_key=%s "
                        "--secret_key=%s get %s %s" %
                        (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey,
                        s3_recover_url,"/data/data-"+str(s3_recover_urls.index(s3_recover_url))+"/redis-bak.tar.gz"))
                status_s3, res_s3 = subprocess.getstatusoutput("s3cmd --no-ssl --region=%s --host=%s "
                                                            "--host-bucket=%s --access_key=%s "
                                                            "--secret_key=%s get %s %s" %
                                                            (S3Region, S3Endpoint, S3Endpoint, S3AccessKey, S3SecretKey,
                                                            s3_recover_url,"/data/data-"+str(s3_recover_urls.index(s3_recover_url))+"/redis-bak.tar.gz"))
                if status_s3 != 0:
                    raise Exception("S3cmd get recover file error %s: %s" % (status_s3, res_s3))
                else:
                    print("S3cmd get recover file succeed: %s" % res_s3)
                    # print("tar -zxvf "+"/data/data-"+str(s3_recover_urls.index(s3_recover_url))+"/redis-bak.tar.gz" + " -C "+ "/data/data-"+str(s3_recover_urls.index(s3_recover_url)) )
                    tar_status, tar_res = subprocess.getstatusoutput("tar -zxvf "+"/data/data-"+str(s3_recover_urls.index(s3_recover_url))+"/redis-bak.tar.gz"+ " -C "+ "/data/data-"+str(s3_recover_urls.index(s3_recover_url)))
                    # print("sed -i 's/\([^ ]*\) [^ ]*/\1 127.0.0.1:6379@16379/' /data/data-"+str(s3_recover_urls.index(s3_recover_url))+ "/node.conf")
                    sed_status, sed_res = subprocess.getstatusoutput("sed -i 's/\([^ ]*\) [^ ]*/\\1 127.0.0.1:6379@16379/' /data/data-"+str(s3_recover_urls.index(s3_recover_url))+ "/node.conf")
                    if tar_status != 0:
                        raise Exception("tar recover file failed %s: %s" % (tar_status, tar_res))
                    else:
                        print("redis recover succeed")
                status_rm, res_rm = subprocess.getstatusoutput("rm -rf "+"/data/data-"+str(s3_recover_urls.index(s3_recover_url))+"/redis-bak.tar.gz")
            return
    except Exception as e:
        print("redis recover error: " + str(e))
        sys.exit(-1)


if __name__ == '__main__':
    redis_recover()
