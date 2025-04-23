from pydantic import BaseModel, Field
from langchain_openai import ChatOpenAI
from langchain_core.output_parsers import JsonOutputParser
from langchain_core.prompts import PromptTemplate
from openai import RateLimitError
import sys, json, os, time

class Tags(BaseModel):
    character: str = Field(description="Character tags are a classification system used to label the multi-dimensional characteristics of fictional characters, including hair color and type, facial features, and body shape. Please generate character settings or analyze tag relevance based on user-provided tags")
    clothing: str = Field(description="Clothing tags is a classification system used to label the clothing worn by fictional characters, including clothing types such as clothes, pants, skirts, and clothing styles such as Chinese style and Japanese style. Please generate clothing settings or analyze tag relevance based on the tags provided by users")
    background: str = Field(description="Background tags are used to mark the background of fictional characters, including specific scene layout and abstract color style. Please generate background settings or analyze tag relevance based on user-provided tags")
    pose: str = Field(description="Pose tags are used to mark the actions of virtual characters, including body movements and facial expressions, but not clothing information and character feature information. Sometimes multiple characters may make body movements at the same time. Please generate pose settings or analyze label relevance based on user-provided labels")

if len(sys.argv) > 2: 
    input_tags = sys.argv[1]
    input_keys = sys.argv[2]
else:
    input_tags = ""
    input_keys = ""

output_parser = JsonOutputParser()
format_instructions = output_parser.get_format_instructions()

os.environ['HTTP_PROXY'] = 'http://127.0.0.1:7890'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:7890'

prompt = PromptTemplate(
    template="Please convert the following tags to JSON format: \n{input_tags}\n{format_instructions}",
    input_variables=["input_tags"],
    partial_variables={"format_instructions": format_instructions}
)

llm = ChatOpenAI(model="gpt-4o-mini", api_key=input_keys)
chain = prompt | llm.with_structured_output(Tags)

data = {}
result = chain.invoke(input_tags).model_dump()
for key, value in result.items():
    if isinstance(value, str):
        data[key] = value.split(',') 
print(json.dumps(data))