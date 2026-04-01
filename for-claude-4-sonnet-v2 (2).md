# AIDLC demo prompts

## Phase 1: Inception

### Step 1.1: Create User Stories

```
Your Role: You are an expert product manager and are tasked with creating well defined user stories that becomes the contract for developing the system as mentioned in the Task section below.

Plan for the work ahead and write your steps in an md file (plan.md) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Your Task: I would like to build an application. In this application, my customers interacts with the products, adding them to the wishlist, then can utilize AI by specifying all their preferences (occasion of usage, styling choices, budget, color combos, etc.). AI understands the preferences and suggests the matchup combos if available, if not then can suggest user to change their choices, recommend new choices from our product database. Once the customer confirms, the system complete the work by prompting the customer to put the combos to cart and make checkout.

Create an /inception/ directory and write the user stories to user_stories.md in the inception directory. Only focus on user stories and nothing else.
```

### Step 1.2: Grouping User Stories into Units

```
Your Role: You are an expert software architect and are tasked with grouping the user stories into multiple units that can be built independently as mentioned in the Task section below.

Plan for the work ahead and write your steps in an md file (plan.md) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Your Task: Refer to the user stories in, /inception/user_stories.md file. Group the user stories into multiple units that can be built independently. Each unit contains highly cohesive user stories that can be built by a single team. The units must be loosely coupled with each other. For each unit, write their respective user stories and acceptance criteria in individual .md files in the /inception/units/ folder. Do not start the technical systems design yet.

Create an integration contract for each of the units, note what API endpoints that each service expose and define the method in an /inception/units/integration_contract.md file.

```

## Phase 2: Construction of one Unit

### Step 2.1: Design Domain Model with DDD

```
Your Role: You are an expert software architect and are tasked with designing the domain model using Domain Driven Design for a unit of of the software system. Refer to the Task section for more details.

Plan for the work ahead and write your steps in an md file (unitname_plan.md, unitname is the unit of work focus to design) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Focus only on the combo portfolio and cart handoff system.

Your Task: Refer to /inception/units/ folder, each md file represents a software unit with the corresponding user stories. Design the Domain Driven Design domain model with all the tactical components including aggregates, entities, value objects, domain events, policies, repositories, domain services etc. Create a new /construction/ folder in the / directory, write the designs details in a /construction/{unit name}/domain_model.md file.

DO NOT generate code snippets.
```

### Step 2.2: Create Logical Design

```
Your Role: You are an expert software architect and are tasked with creating a logical design of a highly scalable, event-driven system according to a Domain Driven Design domain model. Refer to the Task section for more details.

Plan for the work ahead and write your steps in an md file (unitname_plan.md, unitname is the unit of work focus to design) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Focus only on the combo portfolio and cart handoff system.

Your Task: Refer to /construction/{unit name}/domain_model.md file for the domain model and integration document with other services at /inception/units/integration_contract.md file. Generate a logical design for software source code implementation. Write the design document to the /construction/{unit name}/logical_design.md file.

DO NOT generate code snippets.
```

### Step 2.3: Implement Source Code

```
Your Role: You are an expert software engineer and are tasked with implementing a highly scalable, event-driven system according to the Domain Driven Design logical design. Refer to the Task section for more details.

Plan for the work ahead and write your steps in an md file (unitname_plan.md, unitname is the unit of work focus to design) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Focus only on the combo portfolio and cart handoff system.

Your Task: Refer to /construction/{unit name}/logical_design.md file for the logical design details. Generate a very simple and intuitive Golang implementation for the bounded context. Generate the classes in respective individual files but keep them in the /construction/{unit name}/src/ directory based on the proposed file structure. Create a simple demo script that can be run locally to verify the implementation.
```

### Step 2.4: Debugging Source Code

```
Your Role: You are an expert software engineer and are tasked with debugging issues with the demo application.

Resolve the issue below and any other issues to ensure that the demo script can be executed successfully.

Issue:

```

### Step 2.5: Create Tests

```
Your Role: You are an expert quality assurance engineer and are tasked with creating test plans according to the user requirements and technical design for the software systems.

Plan for the work ahead and write your steps in an md file (unitname_plan.md, unitname is the unit of work focus to design) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

Focus only on the combo portfolio and cart handoff system.

Your Task: Refer to /construction/{unit name}/domain_design.md and /construction/{unit name}/logical_design.md files for the software system design. The implementation is in /construction/{unit name}/src directory if you need more details. Refer to the business requirements for this software unit, including user stories and acceptance criteria in individual .md files in the /inception/units/{unit name}.md file. Generate test plans to test the backend system of this software unit.
Use property based testing (PBT) in creating the test plan and test implementation for this project.
```

## Phase 3: Operations of one Unit

### Step 3.1: Create IaC Scripts

```
Your Role: You are an expert devops engineer and are tasked with creating deployment scripts in cloud formation to deploy the selected unit of software to an AWS account based on the specifications of the logical deisgn.

Plan for the work ahead and write your steps in an md file (plan.md) with checkboxes for each step in the plan. If any step needs my clarification, add the questions with the [Question] tag and create an empty [Answer] tag for me to fill the answer. Do not make critical decisions on your own. Upon completing the plan, ask for my review and approval. After my approval, you can go ahead to execute the same plan one step at a time. Once you finish each step, mark the checkboxes as done in the plan.

(REPLACE THIS!) Focus only on the booking and reservation system.

(REVIEW THIS!) Your Task: Refer to /construction/{unit name}/logical_design.md files for the software system design. The implementation is in /construction/{unit name}/src directory if you need more details. Generate deployment scripts in cloud formation in /construction/{unit name}/deploy to deploy the backend system of this software unit to AWS in /operation/{unit name}/ directory. Create an API Gateway for this unit with relevant endpoints and methods according to the /inception/units/integration_contract.md, connect to a lambda function which will containing the backend source code. Tag all AWS resources for this unit with the unit's name.
