"""
搜索工具 - 为Agent提供视频搜索能力
通过调用Go API的搜索接口实现
"""
import logging
from typing import Optional
import httpx
from langchain_core.tools import StructuredTool
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)

GO_API_BASE_URL = "http://api:8000"


class SearchInput(BaseModel):
    """搜索工具输入参数"""
    query: str = Field(
        ...,
        description="搜索关键词，用于搜索视频标题和描述",
        example="搞笑的猫咪视频"
    )
    sort: Optional[str] = Field(
        "relevance",
        description="排序方式：relevance(相关性)/time(时间)/hot(热度)，默认为relevance",
        example="hot"
    )
    page: int = Field(
        1,
        ge=1,
        description="页码，从1开始，默认为1",
        example=1
    )
    page_size: int = Field(
        5,
        ge=1,
        le=10,
        description="每页返回的视频数量，最多10个，默认为5",
        example=5
    )
    author_id: Optional[int] = Field(
        None,
        description="可选：按作者ID筛选",
        example=None
    )


async def _search_videos_func(
    query: str,
    sort: str = "relevance",
    page: int = 1,
    page_size: int = 5,
    author_id: Optional[int] = None
) -> str:
    """
    搜索视频的内部函数
    
    Args:
        query: 搜索关键词
        sort: 排序方式
        page: 页码
        page_size: 每页数量
        author_id: 作者ID（可选）
        
    Returns:
        格式化的搜索结果字符串
    """
    try:
        params = {
            "q": query,
            "sort": sort,
            "page": page,
            "page_size": page_size
        }
        if author_id:
            params["author_id"] = author_id
        
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.get(
                f"{GO_API_BASE_URL}/api/v1/search/videos",
                params=params
            )
            response.raise_for_status()
            result = response.json()
        
        if not result.get("success"):
            return f"搜索失败：{result.get('message', '未知错误')}"
        
        data = result.get("data", {})
        videos = data.get("videos", [])
        total = data.get("total", 0)
        
        if not videos:
            return f"没有找到与「{query}」相关的视频。"
        
        response_parts = [f"找到 {total} 个相关视频，以下是前 {len(videos)} 个：\n"]
        
        for idx, video in enumerate(videos, 1):
            video_info = f"{idx}. 【{video.get('title', '无标题')}】"
            
            if video.get('author_name'):
                video_info += f" by @{video['author_name']}"
            
            stats = []
            if video.get('view_count'):
                stats.append(f"播放量: {video['view_count']}")
            if video.get('favorite_count'):
                stats.append(f"点赞: {video['favorite_count']}")
            if video.get('comment_count'):
                stats.append(f"评论: {video['comment_count']}")
            
            if stats:
                video_info += f" ({', '.join(stats)})"
            
            if video.get('description'):
                desc = video['description'][:50] + "..." if len(video['description']) > 50 else video['description']
                video_info += f"\n   简介: {desc}"
            
            if video.get('play_url'):
                video_info += f"\n   播放链接: {video['play_url']}"
            
            if video.get('cover_url'):
                video_info += f"\n   封面: {video['cover_url']}"
            
            response_parts.append(video_info)
        
        return "\n".join(response_parts)
        
    except httpx.HTTPStatusError as e:
        logger.error(f"搜索API返回错误: {e.response.status_code}")
        return f"搜索服务暂时不可用，请稍后再试。"
    except Exception as e:
        logger.error(f"搜索视频失败: {e}", exc_info=True)
        return f"搜索过程中出现错误：{str(e)}"


def create_search_tool() -> StructuredTool:
    """
    创建搜索工具
    
    Returns:
        StructuredTool: LangChain 搜索工具实例
    """
    return StructuredTool.from_function(
        func=_search_videos_func,
        name="search_videos",
        description="""搜索视频工具。根据用户提供的搜索关键词，在视频标题和描述中查找相关视频。
        
使用场景：
- 用户想要找特定主题的视频（如"搞笑的猫咪"、"Python教程"）
- 用户想要找某个作者发布的视频
- 用户想要找热门或最新发布的视频

参数说明：
- query: 搜索关键词，必填
- sort: 排序方式，可选（relevance/time/hot），默认按相关性排序
- page: 页码，默认第1页
- page_size: 每页数量，最多10个，默认5个

返回格式：
返回匹配的视频列表，包含标题、作者、播放量、点赞数、评论数和简介等信息。""",
        args_schema=SearchInput,
        coroutine=_search_videos_func
    )
