"""
Agent对话API
提供同步和流式对话接口
"""
import json
import logging
from typing import Optional
from fastapi import APIRouter, HTTPException, Header
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from app.agent.service import get_agent_service
from app.agent.context import get_memory_store
from app.core.auth import get_user_id_str

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/v1/agent", tags=["Agent对话"])


class ChatRequest(BaseModel):
    """对话请求"""
    message: str = Field(..., description="用户消息")
    chat_id: str = Field(..., description="对话ID")


class ChatResponse(BaseModel):
    """对话响应"""
    code: int = Field(200, description="状态码")
    message: str = Field("success", description="状态消息")
    data: Optional[dict] = Field(None, description="额外数据")
    ai_reply: Optional[str] = Field(None, description="AI回复")
    chat_id: Optional[str] = Field(None, description="对话ID")


class ChatListResponse(BaseModel):
    """对话列表响应"""
    code: int = Field(200, description="状态码")
    message: str = Field("success", description="状态消息")
    data: list = Field(default_factory=list, description="对话列表")


@router.post("/invoke", response_model=ChatResponse)
async def invoke_chat(
    request: ChatRequest,
    authorization: Optional[str] = Header(None)
):
    """
    同步调用Agent对话接口
    """
    try:
        message = request.message.strip()
        if not message:
            return ChatResponse(
                code=400,
                message="消息不能为空",
                ai_reply=None,
                chat_id=None
            )
        
        if not request.chat_id:
            return ChatResponse(
                code=400,
                message="chat_id不能为空",
                ai_reply=None,
                chat_id=None
            )
        
        user_id = get_user_id_str(authorization)
        chat_id = request.chat_id
        
        memory_store = get_memory_store()
        agent_service = get_agent_service()
        
        if not agent_service.is_available():
            return ChatResponse(
                code=503,
                message="Agent服务暂不可用，请检查配置",
                ai_reply=None,
                chat_id=None
            )
        
        memory_messages = await memory_store.get_records(user_id, chat_id)
        
        memory_messages.append({
            "role": "user",
            "content": message
        })
        
        ai_reply = await agent_service.ainvoke(memory_messages)
        
        memory_messages.append({
            "role": "assistant",
            "content": ai_reply
        })
        
        await memory_store.save_records(user_id, chat_id, memory_messages)
        await memory_store.add_chat_id(user_id, chat_id)
        
        return ChatResponse(
            code=200,
            message="success",
            ai_reply=ai_reply,
            chat_id=chat_id
        )
    except Exception as e:
        logger.error(f"Agent对话失败: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Agent对话失败: {str(e)}")


@router.post("/stream")
async def stream_chat(
    request: ChatRequest,
    authorization: Optional[str] = Header(None)
):
    """
    流式返回Agent对话接口
    """
    try:
        message = request.message.strip()
        if not message:
            async def error_response():
                chunk = {"code": 400, "message": "消息不能为空", "data": None}
                yield f"data: {json.dumps(chunk, ensure_ascii=False)}\n\n"
            
            return StreamingResponse(
                error_response(),
                media_type="text/event-stream"
            )
        
        if not request.chat_id:
            async def error_response():
                chunk = {"code": 400, "message": "chat_id不能为空", "data": None}
                yield f"data: {json.dumps(chunk, ensure_ascii=False)}\n\n"
            
            return StreamingResponse(
                error_response(),
                media_type="text/event-stream"
            )
        
        user_id = get_user_id_str(authorization)
        chat_id = request.chat_id
        
        memory_store = get_memory_store()
        agent_service = get_agent_service()
        
        if not agent_service.is_available():
            async def error_response():
                chunk = {"code": 503, "message": "Agent服务暂不可用", "data": None}
                yield f"data: {json.dumps(chunk, ensure_ascii=False)}\n\n"
            
            return StreamingResponse(
                error_response(),
                media_type="text/event-stream"
            )
        
        memory_messages = await memory_store.get_records(user_id, chat_id)
        
        memory_messages.append({
            "role": "user",
            "content": message
        })
        
        async def generate_response():
            """生成流式响应"""
            full_reply = ""
            
            try:
                async for chunk in agent_service.stream(memory_messages):
                    full_reply += chunk
                    
                    data_chunk = {
                        "code": 200,
                        "message": "streaming",
                        "data": {"content": chunk}
                    }
                    yield f"data: {json.dumps(data_chunk, ensure_ascii=False)}\n\n"
                
                memory_messages.append({
                    "role": "assistant",
                    "content": full_reply
                })
                
                await memory_store.save_records(user_id, chat_id, memory_messages)
                await memory_store.add_chat_id(user_id, chat_id)
                
                end_chunk = {
                    "code": 200,
                    "message": "done",
                    "data": {"chat_id": chat_id}
                }
                yield f"data: {json.dumps(end_chunk, ensure_ascii=False)}\n\n"
            except Exception as e:
                logger.error(f"流式Agent对话失败: {e}", exc_info=True)
                error_chunk = {
                    "code": 500,
                    "message": f"对话过程中出现错误：{str(e)}",
                    "data": None
                }
                yield f"data: {json.dumps(error_chunk, ensure_ascii=False)}\n\n"
        
        return StreamingResponse(
            generate_response(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no"
            }
        )
    except Exception as e:
        logger.error(f"流式Agent对话失败: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"流式Agent对话失败: {str(e)}")


@router.get("/chats", response_model=ChatListResponse)
async def get_chat_list(
    authorization: Optional[str] = Header(None)
):
    """获取用户的对话列表"""
    try:
        user_id = get_user_id_str(authorization)
        memory_store = get_memory_store()
        
        chat_ids = await memory_store.get_chat_list(user_id)
        
        chat_list = []
        for chat_id in chat_ids[:20]:
            preview = await memory_store.get_chat_preview(user_id, chat_id)
            if preview:
                chat_list.append(preview)
        
        return ChatListResponse(
            code=200,
            message="success",
            data=chat_list
        )
    except Exception as e:
        logger.error(f"获取对话列表失败: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"获取对话列表失败: {str(e)}")


@router.get("/chats/{chat_id}")
async def get_chat_messages(
    chat_id: str,
    authorization: Optional[str] = Header(None)
):
    """获取指定对话的完整消息历史"""
    try:
        user_id = get_user_id_str(authorization)
        memory_store = get_memory_store()
        
        messages = await memory_store.get_records(user_id, chat_id)
        
        return {
            "code": 200,
            "message": "success",
            "data": {
                "chat_id": chat_id,
                "messages": messages
            }
        }
    except Exception as e:
        logger.error(f"获取对话消息失败: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"获取对话消息失败: {str(e)}")


@router.delete("/chats/{chat_id}")
async def delete_chat(
    chat_id: str,
    authorization: Optional[str] = Header(None)
):
    """删除指定对话"""
    try:
        user_id = get_user_id_str(authorization)
        memory_store = get_memory_store()
        
        await memory_store.delete_chat_id(user_id, chat_id)
        
        return {"code": 200, "message": "对话已删除"}
    except Exception as e:
        logger.error(f"删除对话失败: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"删除对话失败: {str(e)}")


@router.get("/health")
async def agent_health():
    """Agent服务健康检查"""
    agent_service = get_agent_service()
    memory_store = get_memory_store()
    
    return {
        "agent_available": agent_service.is_available(),
        "memory_available": memory_store.is_available()
    }
