# Fast S3

The fastest S3 sync North of Minnesota.

## Examples

- sync a specific day
```
./_bin/s3.darwin.amd64 sync --config ../fortuna/config/fortuna.yml
```

```
time s3cmd sync   \
    --access_key=$(yq eval '.digital_ocean_s3_access_key_id' ${FORTUNA_CONFIG})     \
    --secret_key=$(yq eval '.digital_ocean_s3_secret_access_key' ${FORTUNA_CONFIG})     \
    --host=$(yq eval '.digital_ocean_s3_endpoint' ${FORTUNA_CONFIG})     \
    --host-bucket='%(bucket)s.sfo2.digitaloceanspaces.com'  \
        s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ \
        _bin/fortuna-stock-data/

real    0m47.689s
user    0m58.611s
sys     0m1.170s
```


```
time /Users/tomcocozzello/go/src/github.com/larrabee/s3sync/s3sync \
    --sk $(yq eval '.digital_ocean_s3_access_key_id' ${FORTUNA_CONFIG}) \
    --ss $(yq eval '.digital_ocean_s3_secret_access_key' ${FORTUNA_CONFIG}) \
    --se sfo2.digitaloceanspaces.com \
        s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ \
        _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/
```










```
time ./_bin/s3.darwin.amd64 sync --config ../fortuna/config/fortuna.yml --s3-prefix TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ --destination-dir _bin/fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/

real    0m10.023s
user    0m0.942s
sys     0m1.312s
```

```
time ./_bin/s3.darwin.amd64 sync --config ../fortuna/config/fortuna.yml --s3-prefix 'TIME_SERIES_INTRADAY_V2/1min/2021-*' --destination-dir _bin/fortuna-stock-data/
```
