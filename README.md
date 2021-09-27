# S3

Dead simple and fast S3 sync

## Getting started

1. Download cli from the [Github release](https://github.com/razzo-lunare/s3/releases/latest)

2. Setup the [s3 config](configs/s3.yaml.example)

3. Test the cli following the examples below

## Source or Destination list

`s3://<BUCKET NAME>/<OPTIONAL PATH ON S3>`: A location inside an S3 bucket.

example with a subpath inside the bucket
```
s3://dates-bucket/2021-02-10/
```

example with only the bucket
```
s3://dates-bucket/
```

`filesystem://<AN ABSOLUTE OR RELATIVE PATH>`: A location on your local filesystem

example relative path
```
filesystem://outputs/dates/
```
example absolute path
```
filesystem:///root/outputs/dates/
```

## Examples

download from s3
```bash
s3 sync \
    --config configs/s3.yaml \
    --source s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/ \
    --destination filesystem://_bin/fortuna-stock-data/
```

upload to s3
```bash
s3 sync \
    --config configs/s3.yaml \
    --source filesystem://_bin/fortuna-stock-data/ \
    --destination s3://fortuna-stock-data-new/TIME_SERIES_INTRADAY_V2/1min/2021-02-10/
```
