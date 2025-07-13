// Example data to send
const readingRequest = {
    test: {
        testNumber: 1,
        sections: [
            {
                sectionNumber: 1,
                timeAllowed: 30,
                passages: [
                    {
                        passageNumber: 1,
                        title: "Sample Passage",
                        content: [
                            {
                                paragraphSummary: "Summary of paragraph 1",
                                keyWords: "key, words",
                                keySentence: "Key sentence of paragraph 1"
                            }
                        ],
                        questions: [
                            {
                                questionNumber: 1,
                                type: "MultipleChoice",
                                content: "What is the main idea?",
                                options: ["Option 1", "Option 2", "Option 3"],
                                correctAnswer: "Option 1"
                            }
                        ]
                    }
                ]
            }
        ]
    }
};

// Function to send the request
async function sendReadingRequest() {
    try {
        const response = await fetch('http://your-backend-url/endpoint', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(readingRequest)
        });

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        const data = await response.json();
        console.log('Success:', data);
    } catch (error) {
        console.error('Error:', error);
    }
}

// Call the function to send the request
sendReadingRequest();
