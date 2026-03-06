---
name: "Model request"
description: Asking to add a model to the list
title: "(short issue description)"
labels: ["model-request"]
assignees: []

body:
    - type: markdown
        attributes:
            value: "## Model Request"
    
    - type: textarea
        id: description
        attributes:
            label: "Model Description"
            description: "Describe the model you'd like to add"
            placeholder: "Provide details about the model..."
        validations:
            required: true
    
    - type: input
        id: model-name
        attributes:
            label: "Model Name"
            description: "Official name of the model"
        validations:
            required: true
    
    - type: textarea
        id: use-case
        attributes:
            label: "Use Case"
            description: "What problem does this model solve?"
        validations:
            required: true
    
    - type: checkboxes
        id: checklist
        attributes:
            label: "Checklist"
            options:
                - label: "Model is open source"
                - label: "Documentation is available"
                - label: "Model has proper license"