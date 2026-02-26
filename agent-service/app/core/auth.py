"""
JWT认证模块 - 解析和验证JWT Token
"""
import logging
import base64
import json
from typing import Optional
from app.core.config import settings

logger = logging.getLogger(__name__)


def decode_jwt_payload(token: str) -> Optional[dict]:
    """
    解码JWT Token的payload部分（不验证签名）
    
    注意：这里只做解码，不做签名验证
    因为签名验证应该由Go API服务完成，Agent服务信任已通过Go API的请求
    
    Args:
        token: JWT Token字符串
        
    Returns:
        解码后的payload字典，失败返回None
    """
    try:
        parts = token.split('.')
        if len(parts) != 3:
            logger.warning("Invalid JWT format: expected 3 parts")
            return None
        
        payload_b64 = parts[1]
        padding = 4 - len(payload_b64) % 4
        if padding != 4:
            payload_b64 += '=' * padding
        
        payload_json = base64.urlsafe_b64decode(payload_b64)
        payload = json.loads(payload_json)
        
        return payload
    except Exception as e:
        logger.warning(f"Failed to decode JWT payload: {e}")
        return None


def get_user_id_from_token(authorization: Optional[str]) -> Optional[int]:
    """
    从Authorization头中提取用户ID
    
    Args:
        authorization: Authorization头的值，格式为 "Bearer <token>"
        
    Returns:
        用户ID，如果解析失败返回None
    """
    if not authorization:
        return None
    
    if not authorization.startswith("Bearer "):
        logger.warning("Invalid authorization header format")
        return None
    
    token = authorization[7:]
    
    payload = decode_jwt_payload(token)
    if not payload:
        return None
    
    user_id = payload.get("user_id")
    if user_id is None:
        logger.warning("user_id not found in JWT payload")
        return None
    
    try:
        return int(user_id)
    except (ValueError, TypeError):
        logger.warning(f"Invalid user_id in JWT payload: {user_id}")
        return None


def get_user_id_str(authorization: Optional[str]) -> str:
    """
    从Authorization头中提取用户ID并返回字符串格式
    
    Args:
        authorization: Authorization头的值
        
    Returns:
        用户ID字符串，如果解析失败返回"anonymous"
    """
    user_id = get_user_id_from_token(authorization)
    if user_id is None:
        return "anonymous"
    return str(user_id)
