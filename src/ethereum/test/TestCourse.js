const { expectEvent, constants, expectRevert } = require('openzeppelin-test-helpers');

const Course = artifacts.require('Course');

contract('Course', accounts => {
    const [teacher, evaluator, student, other] = accounts;
    let course = null;
    const digest = web3.utils.sha3('QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG');

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            course = await Course.new([teacher, evaluator], 2);
            (await course.isOwner(teacher)).should.equal(true);
            (await course.isOwner(evaluator)).should.equal(true);
            assert(course.quorum(), 2);
        });
    });

    describe('CRUD operations', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
        });

        it('should add a new student', async () => {
            const { logs } = await course.addStudent(student, {from: teacher});

            expectEvent.inLogs(logs, 'StudentAdded', {
                student: student,
                requester: teacher
            });

            (await course.enrolled_students(student)).should.equal(true);
        });

        it('should not add a student twice', async () => {
            await course.addStudent(student, {from: teacher});
            await expectRevert(
                course.addStudent(student, {from: evaluator}),
                "Course: student already registered in this course"
            );
        });

        it('should verify if a student is enrolled in the course', async () => {
            (await course.isEnrolled(student)).should.equal(false);
            await course.addStudent(student, {from: teacher});
            (await course.isEnrolled(student)).should.equal(true);
        });

        it('should not allow zero address', async () => {
            await expectRevert(
                course.addStudent(constants.ZERO_ADDRESS, {from: teacher}),
                "Course: student is the zero address"
            );
        });

        it('should allow an owner to remove an enrolled student', async () => {
            await course.addStudent(student, {from: evaluator});
            (await course.enrolled_students(student)).should.equal(true);

            const { logs } = await course.removeStudent(student, {from: teacher});

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: teacher
            });

            (await course.enrolled_students(student)).should.equal(false);
        });

        it('should revert if try to remove an unregistered student', async () => {
            await expectRevert(
                course.removeStudent(student, { from: teacher }),
                "Course: student does not registered in this course"
            );
        });

        it('should allow a student to renounce the course', async () => {
            await course.addStudent(student, {from: teacher});
            (await course.enrolled_students(student)).should.equal(true);

            const { logs } = await course.renounceCourse({from: student});

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: student
            });

            (await course.enrolled_students(student)).should.equal(false);
        });
    });

    describe('Base class operations', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
        });

        it('should issue a credential for a enrolled student', async () => {
            await course.addStudent(student, {from: teacher});
            await course.issue(student, digest, { from: teacher });
        });

        it('should not issue a credential for a non-enrolled address', async () => {
            await expectRevert(
                course.issue(other, digest, { from: teacher }),
                'Course: student not registered'
            );
        });
    });
});
