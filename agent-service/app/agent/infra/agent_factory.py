"""
Agent工厂类 - 创建和管理Agent实例
"""
import logging
from typing import Optional, List, Any
from langchain_core.tools import BaseTool
from langgraph.prebuilt import create_react_agent
from app.agent.infra.llm_factory import get_llm

logger = logging.getLogger(__name__)


def create_agent_graph(
    tools: Optional[List[BaseTool]] = None
) -> Any:
    """
    创建Agent实例（使用LangGraph）
    
    Args:
        tools: 工具列表（可选，如果为None则默认包含搜索工具）
        
    Returns:
        CompiledStateGraph实例（LangGraph Agent）
    """
    llm = get_llm()
    
    if llm is None:
        logger.error("LLM实例不可用，无法创建Agent（请检查DASHSCOPE_API_KEY配置）")
        raise ValueError("LLM服务不可用，请检查配置")
    
    if tools is None:
        from app.agent.tools import create_search_tool
        tools = [create_search_tool()]
    
    try:
        agent = create_react_agent(
            model=llm,
            tools=tools
        )
        
        logger.info(f"✓ Agent已创建 - tools: {len(tools)}")
        return agent
    except Exception as e:
        logger.error(f"创建Agent失败: {e}", exc_info=True)
        raise
