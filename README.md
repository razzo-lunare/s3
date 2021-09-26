# S3

Dead simple and fast S3 sync

## Examples

- sync a specific day
```
./_bin/s3.darwin.amd64 sync --config ../fortuna/config/fortuna.yml
```


TODO clean up the readme

## Performance Report
```
while true; do rm -rf _bin/fortuna-stock-data && time s3cmd sync  --no-preserve \
    --access_key=$(yq eval '.digital_ocean_s3_access_key_id' ${FORTUNA_CONFIG})     \
    --secret_key=$(yq eval '.digital_ocean_s3_secret_access_key' ${FORTUNA_CONFIG})     \
    --host=$(yq eval '.digital_ocean_s3_endpoint' ${FORTUNA_CONFIG})     \
    --host-bucket='%(bucket)s.sfo2.digitaloceanspaces.com'  \
        s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ \
        _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ && sleep 5; done
```
###  529 items
0m48.973s
0m48.928s
0m49.608s

```
# 529 test
time /Users/tomcocozzello/go/src/github.com/larrabee/s3sync/s3sync \
    --sk $(yq eval '.digital_ocean_s3_access_key_id' ${FORTUNA_CONFIG}) \
    --ss $(yq eval '.digital_ocean_s3_secret_access_key' ${FORTUNA_CONFIG}) \
    --se sfo2.digitaloceanspaces.com \
        s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/ \
        _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/

# New test
FORTUNA_CONFIG=configs/s3.yaml
time /Users/tomcocozzello/go/src/github.com/larrabee/s3sync/s3sync \
    --sk $(yq eval '.digital_ocean_s3_access_key_id' ${FORTUNA_CONFIG}) \
    --ss $(yq eval '.digital_ocean_s3_secret_access_key' ${FORTUNA_CONFIG}) \
    --se sfo2.digitaloceanspaces.com \
        s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10 \
        _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/
```
###  529 items

0m5.631s
0m5.022s
0m5.121s
0m5.151s
0m5.180s
0m7.414s
0m5.220s
0m5.442s
0m7.844s
0m6.055s


```
time ./_bin/s3.darwin.amd64 sync --config ../fortuna/config/fortuna.yml --s3-prefix TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ --destination-dir _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/

real    0m10.023s
user    0m0.942s
sys     0m1.312s
```

```
time ./_bin/s3.darwin.amd64 sync \
    --config ../fortuna/config/fortuna.yml \
    --source TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ \
    --destination _bin/fortuna-stock-data/
```

###  529 items
0m5.195s
0m5.561s
0m5.570s
0m5.269s
0m5.093s
0m5.025s
0m5.183s
0m5.178s
0m5.992s
0m5.148s
