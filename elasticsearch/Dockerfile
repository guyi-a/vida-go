# Elasticsearch Dockerfile（含IK中文分词插件）

FROM docker.elastic.co/elasticsearch/elasticsearch:8.11.0

# 安装IK中文分词插件
# 使用 infinilabs 官方 CDN 地址（推荐方式）
RUN bin/elasticsearch-plugin install -b \
    https://get.infini.cloud/elasticsearch/analysis-ik/8.11.0

# 验证插件安装
RUN bin/elasticsearch-plugin list | grep analysis-ik || echo "Warning: IK plugin installation may have failed"
