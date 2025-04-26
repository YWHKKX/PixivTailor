from pydantic import BaseModel, Field
from langchain_openai import ChatOpenAI
from langchain_core.output_parsers import JsonOutputParser
from langchain_core.prompts import PromptTemplate
from openai import RateLimitError
import sys, json, os, time

class Tags(BaseModel):
    character: str = Field(description="Character labeling is a classification system used to label the multidimensional characteristics of a fictional character, including hair color and type, facial features, and body shape but not clothing features. Generate role settings or analyze label affinities based on user-provided labels")
    clothing: str = Field(description="Clothing tags is a classification system used to label the clothing worn by fictional characters, including clothing types such as clothes, pants, skirts, and clothing styles such as Chinese style and Japanese style. Please generate clothing settings or analyze tag relevance based on the tags provided by users")
    background: str = Field(description="Background tags are used to mark the background of fictional characters, including specific scene layout and abstract color style. Please generate background settings or analyze tag relevance based on user-provided tags")
    action: str = Field(description="Action tags are used to mark the virtual character's movements, including body movements, postures, and facial expressions, but do not include any clothing information and character trait information. Generate pose settings based on user-provided labels or analyze label relevance")
    other: str = Field(description="Other tags are used to store uncategorized labels, and if a label does not belong to any of the above label categories, other labels are placed")

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