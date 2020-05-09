const expect = require('@jest/globals').expect
require('./test_report.js')

/**
 * @property {Array.<TestResults>} TestResults
 */
const mockData = [
  {
    "TestResults": [{
      TestName: "my_sample_test 1",
      Package: "test/package 1",
      Output: [
        "test output A 1\n",
        "test output A 2\n",
        "test output A 3\n",
      ]
    }]
  }, {
    "TestResults": [{
      TestName: "my_sample_test 2",
      Package: "test/package 2",
      Output: [
        "test output B 1\n",
        "test output B 2\n",
        "test output B 3\n",
      ]
    }]
  }, {
    "TestResults": [{
      TestName: "my_sample_test 3",
      Package: "test/package 3",
      Output: [
        "test output C 1\n",
        "test output C 2\n",
        "test output C 3\n",
      ]
    }]
  }]


function createTestElements() {
  const testResultsElem = document.createElement('div')
  testResultsElem.id = 'testResults'
  const testGroupListElem = document.createElement('div')
  testGroupListElem.classList.add('cardContainer')
  testGroupListElem.classList.add('testGroupList')
  testGroupListElem.id = 'testGroupList'

  let counter = 0
  mockData.forEach((_) => {
    let testResultGroup = document.createElement('div')
    testResultGroup.id = counter.toString()
    counter += 1
    testResultsElem.insertAdjacentElement("beforeend", testResultGroup)
  })

  return {
    data: mockData,
    testResultsElem: testResultsElem,
    testGroupListElem: testGroupListElem
  }
}

test('test GoTestReport constructor with click event on a test group', () => {
  const testElements = createTestElements()
  const goTestReport = window.GoTestReport(testElements);
  const invocationCounts = {testResultsClickHandler: 0}
  goTestReport.testResultsClickHandler = function (target,
                                                   shiftKey,
                                                   data,
                                                   selectedItems,
                                                   testGroupListHandler) {
    expect(true).toBe(true)
    expect(target.outerHTML).toBe(`<div id="0"></div>`)
    expect(shiftKey).toBe(false)
    expect(data).toBe(mockData)
    expect(selectedItems.testResults).toBeNull()
    expect(selectedItems.selectedTestGroupColor).toBeNull()
    expect(testGroupListHandler).toBe(goTestReport.testGroupListHandler)
    invocationCounts.testResultsClickHandler += 1
  }
  const clickEvent = new MouseEvent('click')
  clickEvent.data = {
    target: testElements.testResultsElem.querySelector('#\\30')
  }
  testElements.testResultsElem.dispatchEvent(clickEvent)
  expect(invocationCounts.testResultsClickHandler).toBe(1)
})

test('test testGroupListHandler', () => {
  const goTestReport = new window.GoTestReport(createTestElements());
  const divElem = document.createElement('div')
  const testIdAttr = document.createAttribute('data-testid')
  testIdAttr.value = "0"
  const indexAttr = document.createAttribute('data-index')
  indexAttr.value = "0"
  divElem.attributes.setNamedItem(testIdAttr)
  divElem.attributes.setNamedItem(indexAttr)
  goTestReport.testGroupListHandler(divElem, mockData)
  const testOutputDiv = divElem.querySelector('div.testOutput')
  expect(testOutputDiv).toBeDefined()

  const consoleElem = testOutputDiv.querySelector('.console.failed')
  expect(consoleElem).toBeDefined()
  expect(consoleElem.textContent).toBe('test output A 1\ntest output A 2\ntest output A 3\n')

  const testDetailElem = testOutputDiv.querySelector('.testDetail')
  expect(testDetailElem).toBeDefined()

  const packageElem = testDetailElem.querySelector('.package')
  expect(packageElem).toBeDefined()
  expect(packageElem.innerHTML).toBe(`<strong>Package:</strong> test/package 1`)

  const filenameElem = testDetailElem.querySelector('.filename')
  expect(filenameElem).toBeDefined()
  expect(filenameElem.innerHTML).toBe(`<strong>Filename:</strong> `)
})