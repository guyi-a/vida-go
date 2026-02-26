"""
配置管理模块
从环境变量读取配置
"""
import os
from typing import Optional
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """应用配置"""
    
    # Elasticsearch
    ELASTICSEARCH_HOSTS: str = "http://elasticsearch:9200"
    ELASTICSEARCH_INDEX_VIDEOS: str = "videos"
    ELASTICSEARCH_TIMEOUT: int = 30
    ELASTICSEARCH_MAX_RETRIES: int = 3
    
    # Redis
    REDIS_URL: str = "redis://redis:6379/0"
    
    # LLM Configuration
    DASHSCOPE_API_KEY: Optional[str] = None
    LLM_BASE_URL: str = "https://dashscope.aliyuncs.com/compatible-mode/v1"
    LLM_MODEL: str = "qwen-max"
    
    # Memory Configuration
    MEMORY_MAX_TOKENS: int = 2000
    
    class Config:
        env_file = ".env"
        extra = "ignore"


# 全局配置实例
_settings: Optional[Settings] = None


def get_settings() -> Settings:
    """获取配置单例"""
    global _settings
    if _settings is None:
        _settings = Settings()
    return _settings


settings = get_settings()
