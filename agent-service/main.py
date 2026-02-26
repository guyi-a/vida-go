"""
Python Agent服务入口
提供AI Agent功能的HTTP API
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)

# 创建FastAPI应用
app = FastAPI(
    title="Vida-Go Agent Service",
    description="AI Agent service for Vida-Go",
    version="0.1.0",
)

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
async def root():
    """根路径"""
    return {
        "message": "Vida-Go Agent Service",
        "version": "0.1.0",
        "status": "running",
    }


@app.get("/health")
async def health_check():
    """健康检查接口"""
    from app.agent.service import get_agent_service
    from app.agent.context import get_memory_store
    
    agent_service = get_agent_service()
    memory_store = get_memory_store()
    
    return {
        "status": "ok",
        "service": "agent-service",
        "version": "0.1.0",
        "agent_available": agent_service.is_available(),
        "memory_available": memory_store.is_available(),
    }


# 注册Agent路由
from app.api.agent import router as agent_router
app.include_router(agent_router)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
