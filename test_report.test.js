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
      ],
      TestFileName: "test_test.go",
      TestFunctionDetail: {
        Line: 1,
        Col: 10,
      },
    }]
  }, {
    "TestResults": [{
      TestName: "my_sample_test 2",
      Package: "test/package 2",
      Passed: true,
      Output: [
        "test output B 1\n",
        "test output B 2\n",
        "test output B 3\n",
      ],
      TestFileName: "test_test_1.go",
      TestFunctionDetail: {
        Line: 20,
        Col: 1,
      },
    }, {
      TestName: "my_sample_test 3",
      Package: "test/package 3",
      Passed: false,
      Output: [
        "test output C 1\n",
        "test output C 2\n",
        "test output C 3\n",
      ],
      TestFileName: "test_test_2.go",
      TestFunctionDetail: {
        Line: 33,
        Col: 7,
      },
    }]
  }, {
    "TestResults": [{
      TestName: "my_sample_test 4",
      Package: "test/package 4",
      Passed: true,
      Output: [
        "test output D 1\n",
        "test output D 2\n",
        "test output D 3\n",
      ],
      TestFileName: "test_test_3.go",
      TestFunctionDetail: {
        Line: 101,
        Col: 9,
      },
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


/**
 * Returns an element using the provided testGroupId and testIndex.
 * @param {number} testGroupId
 * @param {number} testIndex
 * @returns {HTMLDivElement}
 */
function createDataGroupElement(testGroupId, testIndex) {
  const divElem = document.createElement('div')
  const testGroupIdAttr = document.createAttribute('data-groupid')
  testGroupIdAttr.value = testGroupId.toString()
  const indexAttr = document.createAttribute('data-index')
  indexAttr.value = testIndex.toString()
  divElem.attributes.setNamedItem(testGroupIdAttr)
  divElem.attributes.setNamedItem(indexAttr)
  return divElem
}

test('test testGroupListHandler using [test group: 0]', () => {
  const goTestReport = new window.GoTestReport(createTestElements());
  let divElem = createDataGroupElement(0, 0)
  goTestReport.testGroupListHandler(divElem, mockData)
  const testOutputDiv = divElem.querySelector('div.testOutput')
  const consoleElem = testOutputDiv.querySelector('.console.failed')
  expect(consoleElem.textContent).toBe('test output A 1\ntest output A 2\ntest output A 3\n')
  const testDetailElem = testOutputDiv.querySelector('.testDetail')
  const packageElem = testDetailElem.querySelector('.package')
  expect(packageElem.innerHTML).toBe(`<strong>Package:</strong> test/package 1`)
  const filenameElem = testDetailElem.querySelector('.filename')
  expect(filenameElem.innerHTML).toBe(`<strong>Filename:</strong> test_test.go &nbsp;&nbsp;<strong>Line:</strong> 1 <strong>Col:</strong> 10`)
})

test('test testGroupListHandler using [test group: 1]', () => {
  const goTestReport = new window.GoTestReport(createTestElements());
  let divElem = createDataGroupElement(2, 0)
  goTestReport.testGroupListHandler(divElem, mockData)
  const testOutputDiv = divElem.querySelector('div.testOutput')
  const consoleElem = testOutputDiv.querySelector('.console')
  expect(consoleElem.textContent).toBe('test output D 1\ntest output D 2\ntest output D 3\n')
  const testDetailElem = testOutputDiv.querySelector('.testDetail')
  const packageElem = testDetailElem.querySelector('.package')
  expect(packageElem.innerHTML).toBe(`<strong>Package:</strong> test/package 4`)
  const filenameElem = testDetailElem.querySelector('.filename')
  expect(filenameElem.innerHTML).toBe(`<strong>Filename:</strong> test_test_3.go &nbsp;&nbsp;<strong>Line:</strong> 101 <strong>Col:</strong> 9`)
})