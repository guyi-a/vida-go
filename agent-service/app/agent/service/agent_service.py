"""
Agent服务 - 使用LangGraph封装Agent调用
支持异步和流式调用
"""
import logging
from pathlib import Path
from typing import List, Dict, Optional, AsyncIterator, Any
from langchain_core.messages import HumanMessage, SystemMessage, AIMessage, BaseMessage
from app.agent.infra.agent_factory import create_agent_graph

logger = logging.getLogger(__name__)


class AgentService:
    """Agent服务类 - 封装LangGraph Agent调用"""
    
    def __init__(self, tools: Optional[List] = None, prompt_file: Optional[str] = None):
        """
        初始化Agent服务
        
        Args:
            tools: 工具列表（可选）
            prompt_file: 提示词文件路径（可选）
        """
        self.system_prompt = self._load_prompt(prompt_file)
        self.agent = None
        
        try:
            self.agent = create_agent_graph(tools=tools)
            logger.info(f"✓ Agent服务已初始化")
        except Exception as e:
            logger.error(f"Agent服务初始化失败: {e}")
    
    def _load_prompt(self, prompt_file: Optional[str] = None) -> str:
        """加载提示词文件"""
        if prompt_file is None:
            prompt_file = "app/agent/prompts/search_agent_prompt.md"
        
        try:
            current_dir = Path(__file__).parent
            project_root = current_dir.parent.parent.parent
            prompt_path = project_root / prompt_file
            
            if prompt_path.exists():
                with open(prompt_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                logger.info(f"✓ 已加载提示词文件: {prompt_file}")
                return content
            else:
                logger.warning(f"提示词文件不存在: {prompt_path}")
                return ""
        except Exception as e:
            logger.error(f"加载提示词文件失败: {e}")
            return ""
    
    def _convert_messages(self, messages: List[Dict]) -> List[BaseMessage]:
        """将消息字典列表转换为LangChain Message对象"""
        langchain_messages = []
        
        has_system_message = any(msg.get("role") == "system" for msg in messages)
        
        if not has_system_message and self.system_prompt:
            langchain_messages.append(SystemMessage(content=self.system_prompt))
        
        for msg in messages:
            role = msg.get("role", "")
            content = msg.get("content", "")
            
            if role == "system":
                langchain_messages.append(SystemMessage(content=content))
            elif role == "user":
                langchain_messages.append(HumanMessage(content=content))
            elif role == "assistant":
                langchain_messages.append(AIMessage(content=content))
        
        return langchain_messages
    
    async def ainvoke(self, messages: List[Dict], **kwargs: Any) -> str:
        """
        异步非流式调用Agent
        
        Args:
            messages: 消息列表
            **kwargs: 其他参数
            
        Returns:
            Agent回复文本
        """
        if not self.agent:
            return "抱歉，Agent服务暂不可用，请检查配置。"
        
        try:
            langchain_messages = self._convert_messages(messages)
            
            if not langchain_messages:
                return "请输入您的问题。"
            
            agent_input = {"messages": langchain_messages}
            agent_input.update(kwargs)
            
            result = await self.agent.ainvoke(agent_input)
            
            if isinstance(result, dict):
                if "messages" in result:
                    messages_list = result["messages"]
                    for msg in reversed(messages_list):
                        if isinstance(msg, AIMessage):
                            return msg.content.strip() if msg.content else ""
                output = result.get("output", "")
                if output:
                    return output.strip()
                return str(result).strip()
            elif isinstance(result, str):
                return result.strip()
            else:
                return str(result).strip()
                
        except Exception as e:
            logger.error(f"Agent调用失败: {e}", exc_info=True)
            return f"抱歉，对话过程中出现错误：{str(e)}"
    
    async def stream(self, messages: List[Dict], **kwargs: Any) -> AsyncIterator[str]:
        """
        流式调用Agent
        
        Args:
            messages: 消息列表
            **kwargs: 其他参数
            
        Yields:
            Agent回复的文本片段
        """
        if not self.agent:
            yield "抱歉，Agent服务暂不可用，请检查配置。"
            return
        
        try:
            langchain_messages = self._convert_messages(messages)
            
            if not langchain_messages:
                yield "请输入您的问题。"
                return
            
            agent_input = {"messages": langchain_messages}
            agent_input.update(kwargs)
            
            full_output = ""
            input_message_count = len(langchain_messages)
            
            async for state in self.agent.astream(agent_input, stream_mode="values"):
                try:
                    if isinstance(state, dict) and "messages" in state:
                        messages_list = state["messages"]
                        new_messages = messages_list[input_message_count:]
                        
                        for msg in reversed(new_messages):
                            if isinstance(msg, AIMessage) and msg.content:
                                current_content = str(msg.content)
                                if len(current_content) > len(full_output):
                                    new_content = current_content[len(full_output):]
                                    full_output = current_content
                                    if new_content:
                                        yield new_content
                                break
                            
                except Exception as e:
                    logger.warning(f"处理流式chunk时出错: {e}")
                    continue
                        
        except Exception as e:
            logger.error(f"Agent流式调用失败: {e}", exc_info=True)
            yield f"抱歉，对话过程中出现错误：{str(e)}"
    
    def is_available(self) -> bool:
        """检查Agent服务是否可用"""
        return self.agent is not None


_agent_service_instance: Optional[AgentService] = None


def get_agent_service(tools: Optional[List] = None) -> AgentService:
    """获取Agent服务单例"""
    global _agent_service_instance
    if _agent_service_instance is None:
        _agent_service_instance = AgentService(tools=tools)
    return _agent_service_instance
